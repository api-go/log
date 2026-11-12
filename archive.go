package log

// The archive payload is versioned binary data inside independent gzip members.
// It is intentionally decoded only through this package, not through zcat.
import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/api-go/cast"
	"github.com/api-go/file"
	"github.com/api-go/rand"
)

var archiveCompressors = sync.Pool{New: func() any { writer, _ := gzip.NewWriterLevel(io.Discard, gzip.BestSpeed); return writer }}

const archiveMagic = "APIGOLG1"
const dictLimit = 16384
const dictByteLimit = 8 << 20
const maxArchiveLine = 16 << 20
const maxArchiveBlock = 32 << 20
const archiveExtraSize = 36

type archiveDictionary struct {
	Identity []byte
	Meta     map[string][]MetaField
	Values   []string
	ids      map[string]uint32
	seen     map[string]bool
	end      int64
	size     int
	seenSize int
}

func dictionaryPath(path string) string { return strings.TrimSuffix(path, ".gz") + ".dict" }

func loadDictionary(path string, repair bool) (*archiveDictionary, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.Size() > 100<<20 {
		return nil, errors.New("log dictionary exceeds file size limit")
	}
	data, err := file.ReadBytes(path)
	if err != nil {
		return nil, err
	}
	if len(data) < 28 || string(data[:8]) != archiveMagic {
		return nil, errors.New("invalid log dictionary header")
	}
	n := int(binary.LittleEndian.Uint32(data[24:28]))
	if n > maxArchiveBlock || n > len(data)-32 || n < 0 {
		return nil, errors.New("incomplete log dictionary metadata")
	}
	if crc32.ChecksumIEEE(data[28:28+n]) != binary.LittleEndian.Uint32(data[28+n:32+n]) {
		return nil, errors.New("log dictionary metadata checksum mismatch")
	}
	d := &archiveDictionary{Identity: bytes.Clone(data[8:24]), ids: map[string]uint32{}, seen: map[string]bool{}}
	if err = cast.UnmarshalJSON(data[28:28+n], &d.Meta); err != nil {
		return nil, err
	}
	pos := 32 + n
	for pos < len(data) {
		if len(d.Values) >= dictLimit {
			return nil, errors.New("log dictionary exceeds limit")
		}
		if len(data)-pos < 4 {
			break
		}
		size := int(binary.LittleEndian.Uint32(data[pos : pos+4]))
		if size > 4096 {
			return nil, errors.New("invalid dictionary entry size")
		}
		if len(data)-pos < 8+size {
			break
		}
		value := data[pos+4 : pos+4+size]
		if crc32.ChecksumIEEE(value) != binary.LittleEndian.Uint32(data[pos+4+size:pos+8+size]) {
			return nil, errors.New("dictionary entry checksum mismatch")
		}
		d.size += len(value)
		if d.size > dictByteLimit {
			return nil, errors.New("log dictionary exceeds byte budget")
		}
		d.ids[string(value)] = uint32(len(d.Values))
		d.Values = append(d.Values, string(value))
		pos += 8 + size
	}
	d.end = int64(pos)
	if pos != len(data) && repair {
		if err = os.Truncate(path, int64(pos)); err != nil {
			return nil, err
		}
	}
	return d, nil
}

func createDictionary(path string) (*archiveDictionary, error) {
	metaLock.RLock()
	meta, err := cast.ToJSONBytes(metaRegistry)
	metaLock.RUnlock()
	if err != nil {
		return nil, err
	}
	identity := rand.Bytes(16)
	data := append([]byte(archiveMagic), identity...)
	data = binary.LittleEndian.AppendUint32(data, uint32(len(meta)))
	data = append(data, meta...)
	data = binary.LittleEndian.AppendUint32(data, crc32.ChecksumIEEE(meta))
	// Exclusive creation avoids silently replacing a dictionary belonging to old data.
	fp, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return nil, err
	}
	_, err = fp.Write(data)
	if err == nil {
		err = fp.Sync()
	}
	closeErr := fp.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		return nil, err
	}
	return loadDictionary(path, false)
}

func (d *archiveDictionary) encode(line []byte, additions *bytes.Buffer) []byte {
	var out []byte
	literal := func(p []byte) {
		if len(p) > 0 {
			out = append(out, 0)
			out = binary.LittleEndian.AppendUint32(out, uint32(len(p)))
			out = append(out, p...)
		}
	}
	start := 0
	for i := 0; i < len(line); i++ {
		if line[i] != '"' {
			continue
		}
		end := i + 1
		for end < len(line) {
			if line[end] == '\\' {
				end += 2
				continue
			}
			if line[end] == '"' {
				end++
				break
			}
			end++
		}
		if end > len(line) {
			break
		}
		value := string(line[i:end])
		key, ok := d.ids[value]
		if !ok && len(value) >= 8 && len(value) <= 4096 && len(d.Values) < dictLimit && d.size+len(value) <= dictByteLimit {
			if d.seen[value] {
				key = uint32(len(d.Values))
				d.Values = append(d.Values, value)
				d.size += len(value)
				d.ids[value] = key
				ok = true
				delete(d.seen, value)
				d.seenSize -= len(value)
				_ = binary.Write(additions, binary.LittleEndian, uint32(len(value)))
				additions.WriteString(value)
				_ = binary.Write(additions, binary.LittleEndian, crc32.ChecksumIEEE([]byte(value)))
			} else if len(d.seen) < dictLimit && d.seenSize+len(value) <= dictByteLimit {
				d.seen[value] = true
				d.seenSize += len(value)
			}
		}
		if ok {
			literal(line[start:i])
			out = append(out, 1)
			out = binary.LittleEndian.AppendUint32(out, key)
			start = end
		}
		i = end - 1
	}
	literal(line[start:])
	return out
}

func (d *archiveDictionary) decode(data []byte) ([]byte, error) {
	var out []byte
	for len(data) > 0 {
		if len(data) < 5 {
			return nil, errors.New("truncated archive token")
		}
		kind, n := data[0], binary.LittleEndian.Uint32(data[1:5])
		data = data[5:]
		switch kind {
		case 0:
			if uint64(n) > uint64(len(data)) {
				return nil, io.ErrUnexpectedEOF
			}
			out = append(out, data[:n]...)
			data = data[n:]
		case 1:
			if uint64(n) >= uint64(len(d.Values)) {
				return nil, errors.New("missing log dictionary entry")
			}
			out = append(out, d.Values[n]...)
		default:
			return nil, errors.New("unknown archive token")
		}
		if len(out) > maxArchiveLine {
			return nil, errors.New("decoded log line exceeds 16 MiB")
		}
	}
	return out, nil
}

type archiveBlock struct {
	offset int64
	size   int64
	lines  uint32
}

// scanArchive walks member headers without decompressing historical log payloads.
func scanArchive(fp *os.File, identity []byte) ([]archiveBlock, int64, error) {
	return scanArchiveFrom(fp, identity, 0)
}

func scanArchiveFrom(fp *os.File, identity []byte, pos int64) ([]archiveBlock, int64, error) {
	stat, err := fp.Stat()
	if err != nil {
		return nil, pos, err
	}
	var blocks []archiveBlock
	header := make([]byte, 12+archiveExtraSize)
	for pos < stat.Size() {
		if stat.Size()-pos < int64(len(header)) {
			return blocks, pos, io.ErrUnexpectedEOF
		}
		if _, err = fp.ReadAt(header, pos); err != nil {
			return blocks, pos, err
		}
		if !bytes.Equal(header[:4], []byte{31, 139, 8, 4}) || binary.LittleEndian.Uint16(header[10:12]) != archiveExtraSize || string(header[12:20]) != archiveMagic {
			return blocks, pos, errors.New("invalid log archive member")
		}
		if !bytes.Equal(header[20:36], identity) {
			return blocks, pos, errors.New("log archive/dictionary identity mismatch")
		}
		size := int64(binary.LittleEndian.Uint64(header[36:44]))
		lines := binary.LittleEndian.Uint32(header[44:48])
		if size < int64(len(header)+8) || size > maxArchiveBlock || lines == 0 {
			return blocks, pos, errors.New("invalid log archive block size")
		}
		if size > stat.Size()-pos {
			return blocks, pos, io.ErrUnexpectedEOF
		}
		blocks = append(blocks, archiveBlock{pos, size, lines})
		pos += size
	}
	return blocks, pos, nil
}

func writeArchiveBlock(fp, dictFP *os.File, d *archiveDictionary, lines [][]byte) error {
	var payload, additions bytes.Buffer
	for _, line := range lines {
		encoded := d.encode(line, &additions)
		_ = binary.Write(&payload, binary.LittleEndian, uint32(len(encoded)))
		payload.Write(encoded)
	}
	if additions.Len() > 0 {
		if _, err := dictFP.Write(additions.Bytes()); err != nil {
			return err
		}
		if err := dictFP.Sync(); err != nil {
			return err
		}
	}
	var compressed bytes.Buffer
	gz := archiveCompressors.Get().(*gzip.Writer)
	gz.Reset(&compressed)
	defer func() { gz.Reset(io.Discard); archiveCompressors.Put(gz) }()
	gz.Header.Extra = append([]byte(archiveMagic), d.Identity...)
	gz.Header.Extra = binary.LittleEndian.AppendUint64(gz.Header.Extra, 0)
	gz.Header.Extra = binary.LittleEndian.AppendUint32(gz.Header.Extra, uint32(len(lines)))
	if _, err := gz.Write(payload.Bytes()); err != nil {
		return err
	}
	if err := gz.Close(); err != nil {
		return err
	}
	data := compressed.Bytes()
	binary.LittleEndian.PutUint64(data[36:44], uint64(len(data)))
	_, err := fp.Write(data)
	return err
}

func readArchiveBlock(fp *os.File, b archiveBlock, d *archiveDictionary, emit func([]byte) error) error {
	reader, err := gzip.NewReader(io.NewSectionReader(fp, b.offset, b.size))
	if err != nil {
		return err
	}
	defer reader.Close()
	payload, err := io.ReadAll(io.LimitReader(reader, maxArchiveBlock+1))
	if err != nil {
		return err
	}
	if len(payload) > maxArchiveBlock {
		return errors.New("log block exceeds size limit")
	}
	for i := uint32(0); i < b.lines; i++ {
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
		if err = emit(line); err != nil {
			return err
		}
	}
	if len(payload) != 0 {
		return fmt.Errorf("unexpected trailing log payload")
	}
	return nil
}
