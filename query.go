package log

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/api-go/file"
	"github.com/api-go/shell"
	"github.com/api-go/timer"
)

// QueryOptions selects logical lines. Head and Tail are mutually exclusive.
// Zero selects the complete file. Follow waits for new complete blocks.
type QueryOptions struct {
	Head   int
	Tail   int
	Follow bool
	JSON   bool
	Raw    bool
	Color  bool
}

// LogPath resolves the current configured file including split and compression.
func LogPath(conf Config, now time.Time) (string, error) {
	if conf.File == "" || conf.File == "console" || strings.Contains(conf.File, "://") {
		return "", errors.New("file logging is disabled")
	}
	path := conf.File
	if conf.SplitTag != "" {
		path += "." + now.Format(conf.SplitTag)
	}
	if !conf.NoCompress && !strings.HasSuffix(path, ".gz") {
		path += ".gz"
	}
	return path, nil
}

// DefaultConfig returns the default logger's effective configuration.
func DefaultConfig() Config { return DefaultLogger.config }

func formatQueryLine(line []byte, opts QueryOptions, meta map[string][]MetaField) string {
	if opts.Raw {
		return string(line)
	}
	get := GetMeta
	if meta != nil {
		get = func(name string) []MetaField { return meta[name] }
	}
	if opts.JSON {
		return toJSON(string(line), get)
	}
	value := viewable(string(line), get)
	if !opts.Color {
		value = shell.StripANSI(value)
	}
	return value
}

// FormatStream formats uncompressed log lines without owning either stream.
func FormatStream(ctx context.Context, input io.Reader, output io.Writer, opts QueryOptions) error {
	scanner := bufio.NewScanner(input)
	scanner.Buffer(make([]byte, 64*1024), maxArchiveLine)
	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			return err
		}
		if len(scanner.Bytes()) == 0 {
			continue
		}
		if _, err := fmt.Fprintln(output, formatQueryLine(scanner.Bytes(), opts, nil)); err != nil {
			return err
		}
	}
	return scanner.Err()
}

var stopQuery = errors.New("log query complete")

// QueryFile restores gzip+dictionary or legacy plaintext. Finite queries report
// incomplete tails; Follow waits for an in-progress final member to complete.
func QueryFile(ctx context.Context, path string, output io.Writer, opts QueryOptions) error {
	return query(ctx, func() string { return path }, output, opts)
}

// Query follows the application's configured log, including date rotation.
func Query(ctx context.Context, conf Config, output io.Writer, opts QueryOptions) error {
	if _, err := LogPath(conf, time.Now()); err != nil {
		return err
	}
	return query(ctx, func() string { p, _ := LogPath(conf, time.Now()); return p }, output, opts)
}

func query(ctx context.Context, resolve func() string, output io.Writer, opts QueryOptions) error {
	if opts.Head < 0 || opts.Tail < 0 || opts.Head > 0 && opts.Tail > 0 {
		return errors.New("head and tail require positive counts and cannot be combined")
	}
	if opts.Head > 0 && opts.Follow {
		return errors.New("head cannot be combined with follow")
	}
	var wheel *timer.TimeWheel
	if opts.Follow {
		wheel = timer.NewTimeWheel(50*time.Millisecond, 20)
		wheel.Start()
		defer wheel.Stop()
	}
	var cachedDictionary *archiveDictionary
	var dictionaryInfo os.FileInfo
	var previousInfo os.FileInfo
	path := ""
	var identity string
	var offset int64
	first := true
	rotationPending := false
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		current := resolve()
		if current != path && path != "" && opts.Follow {
			if !file.Exists(current) {
				current = path
				rotationPending = false
			} else if !rotationPending {
				current = path
				rotationPending = true
			} else {
				rotationPending = false
			}
		}
		if current != path {
			initial := path == ""
			cachedDictionary = nil
			dictionaryInfo = nil
			previousInfo = nil
			path = current
			offset = 0
			identity = ""
			first = initial
		}
		fp, err := file.Open(path)
		if os.IsNotExist(err) && opts.Follow {
			err = nil
		} else if err == nil {
			if info, e := fp.Stat(); e != nil {
				fp.Close()
				return e
			} else {
				if previousInfo != nil && !os.SameFile(previousInfo, info) {
					offset = 0
					first = !opts.Follow
					cachedDictionary = nil
				}
				previousInfo = info
			}
			if strings.HasSuffix(path, ".gz") {
				var d *archiveDictionary
				info, e := os.Stat(dictionaryPath(path))
				if e != nil {
					err = e
				} else if cachedDictionary == nil || dictionaryInfo == nil || !os.SameFile(info, dictionaryInfo) || info.Size() != dictionaryInfo.Size() || !info.ModTime().Equal(dictionaryInfo.ModTime()) {
					cachedDictionary, err = loadDictionary(dictionaryPath(path), false)
					if err == nil {
						dictionaryInfo = info
					}
				}
				d = cachedDictionary
				if err == nil {
					if identity != "" && identity != string(d.Identity) {
						offset = 0
						first = !opts.Follow
					}
					identity = string(d.Identity)
					if stat, e := fp.Stat(); e == nil && stat.Size() < offset {
						offset = 0
						first = !opts.Follow
					}
					var blocks []archiveBlock
					var end int64
					blocks, end, err = scanArchiveFrom(fp, d.Identity, offset)
					if opts.Follow && errors.Is(err, io.ErrUnexpectedEOF) {
						err = nil
					}
					if end < offset {
						offset = 0
						first = !opts.Follow
					}
					// Data block snapshot precedes the second dictionary snapshot:
					// writers persist dictionary additions before exposing data references.
					if err == nil && len(blocks) > 0 {
						latest, e := os.Stat(dictionaryPath(path))
						if e != nil {
							err = e
						} else if latest.Size() != dictionaryInfo.Size() || !latest.ModTime().Equal(dictionaryInfo.ModTime()) {
							d, err = loadDictionary(dictionaryPath(path), false)
							if err == nil && string(d.Identity) != identity {
								err = errors.New("log archive/dictionary identity changed during query")
							}
							if err == nil {
								cachedDictionary = d
								dictionaryInfo = latest
							}
						}
					}
					if err == nil {
						var total int64
						for _, b := range blocks {
							if b.offset >= offset {
								total += int64(b.lines)
							}
						}
						var skip int64
						if first && opts.Tail > 0 && total > int64(opts.Tail) {
							skip = total - int64(opts.Tail)
						}
						emitted := 0
						for _, b := range blocks {
							if b.offset < offset {
								continue
							}
							if skip >= int64(b.lines) {
								skip -= int64(b.lines)
								offset = b.offset + b.size
								continue
							}
							err = readArchiveBlock(fp, b, d, func(line []byte) error {
								if skip > 0 {
									skip--
									return nil
								}
								if e := ctx.Err(); e != nil {
									return e
								}
								if _, e := fmt.Fprintln(output, formatQueryLine(line, opts, d.Meta)); e != nil {
									return e
								}
								emitted++
								if opts.Head > 0 && emitted >= opts.Head {
									return stopQuery
								}
								return nil
							})
							if err != nil {
								break
							}
							offset = b.offset + b.size
						}
					}
				}
			} else {
				offset, err = queryPlain(ctx, fp, output, opts, offset, first)
			}
			closeErr := fp.Close()
			if err == nil {
				err = closeErr
			}
			first = false
		}
		if errors.Is(err, stopQuery) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		if !opts.Follow {
			return nil
		}
		ready := make(chan struct{}, 1)
		wheel.AfterFunc(200*time.Millisecond, func() { ready <- struct{}{} })
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ready:
		}
	}
}

func queryPlain(ctx context.Context, fp *os.File, output io.Writer, opts QueryOptions, offset int64, first bool) (int64, error) {
	stat, err := fp.Stat()
	if err != nil {
		return offset, err
	}
	if stat.Size() < offset {
		offset = 0
		first = !opts.Follow
	}
	if _, err = fp.Seek(offset, io.SeekStart); err != nil {
		return offset, err
	}
	reader := bufio.NewReaderSize(fp, 64*1024)
	var lines []string
	count := 0
	for {
		if err = ctx.Err(); err != nil {
			return offset, err
		}
		// ReadSlice bounds memory even for malformed plaintext inputs.
		var line []byte
		for {
			part, e := reader.ReadSlice('\n')
			line = append(line, part...)
			if len(line) > maxArchiveLine {
				return offset, errors.New("log line exceeds 16 MiB")
			}
			err = e
			if !errors.Is(e, bufio.ErrBufferFull) {
				break
			}
		}
		if err != nil && !errors.Is(err, io.EOF) {
			return offset, err
		}
		if errors.Is(err, io.EOF) && (opts.Follow || len(line) == 0) {
			break
		}
		offset += int64(len(line))
		count++
		text := strings.TrimSuffix(string(line), "\n")
		if first && opts.Tail > 0 {
			if len(lines) < opts.Tail {
				lines = append(lines, text)
			} else {
				lines[(count-1)%opts.Tail] = text
			}
		} else {
			if _, err = fmt.Fprintln(output, formatQueryLine([]byte(text), opts, nil)); err != nil {
				return offset, err
			}
			if opts.Head > 0 && count >= opts.Head {
				return offset, stopQuery
			}
		}
		if errors.Is(err, io.EOF) {
			break
		}
	}
	start := 0
	if opts.Tail > 0 && count > opts.Tail {
		start = count % opts.Tail
	}
	for i := 0; i < len(lines); i++ {
		if _, err = fmt.Fprintln(output, formatQueryLine([]byte(lines[(start+i)%len(lines)]), opts, nil)); err != nil {
			return offset, err
		}
	}
	return offset, nil
}

// FormatArchiveStream restores a compressed stdin stream using an explicit
// dictionary. Unlike QueryFile this sequential stream cannot seek for tail.
func FormatArchiveStream(ctx context.Context, input io.Reader, output io.Writer, dictionary string, opts QueryOptions) error {
	d, err := loadDictionary(dictionary, false)
	if err != nil {
		return err
	}
	reader := bufio.NewReader(input)
	for {
		header, err := reader.Peek(12 + archiveExtraSize)
		if errors.Is(err, io.EOF) && len(header) == 0 {
			return nil
		}
		if err != nil {
			return err
		}
		if string(header[12:20]) != archiveMagic || !bytes.Equal(header[20:36], d.Identity) {
			return errors.New("log stream/dictionary identity mismatch")
		}
		size := binary.LittleEndian.Uint64(header[36:44])
		if size < 56 || size > maxArchiveBlock {
			return errors.New("invalid archive stream member size")
		}
		count := binary.LittleEndian.Uint32(header[44:48])
		data := make([]byte, int(size))
		if _, err = io.ReadFull(reader, data); err != nil {
			return err
		}
		gz, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return err
		}
		payload, err := io.ReadAll(io.LimitReader(gz, maxArchiveBlock+1))
		gz.Close()
		if err != nil {
			return err
		}
		if len(payload) > maxArchiveBlock {
			return errors.New("archive stream block too large")
		}
		for i := uint32(0); i < count; i++ {
			if err = ctx.Err(); err != nil {
				return err
			}
			if len(payload) < 4 {
				return io.ErrUnexpectedEOF
			}
			n := binary.LittleEndian.Uint32(payload)
			payload = payload[4:]
			if uint64(n) > uint64(len(payload)) {
				return io.ErrUnexpectedEOF
			}
			line, err := d.decode(payload[:n])
			if err != nil {
				return err
			}
			payload = payload[n:]
			if _, err = fmt.Fprintln(output, formatQueryLine(line, opts, d.Meta)); err != nil {
				return err
			}
		}
		if len(payload) > 0 {
			return errors.New("unexpected archive stream trailing payload")
		}
	}
}
