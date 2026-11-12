package log

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/api-go/cast"
)

// RunCommand implements log and logv without global flags, process exit or
// starting application services. Flags may appear before or after filenames.
func RunCommand(ctx context.Context, name string, args []string, input io.Reader, output io.Writer) error {
	_ = LoadMeta(metaFilePath) // Retain legacy stdin/plaintext custom-model formatting.
	opts := QueryOptions{}
	var dictionary string
	var paths []string
	var selected, followSet bool
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			paths = append(paths, args[i+1:]...)
			break
		}
		switch arg {
		case "-h", "-help", "--help":
			_, err := fmt.Fprintln(output, "Usage: log [-tail N|-head N|-cat] [-f] [-json|-raw] [files ...]\n       logv [-json|-raw] [files ...]\nlog defaults to the configured file, -tail 100 -f. logv without files formats stdin.")
			return err
		case "-config", "--config":
			if i+1 >= len(args) {
				return errors.New("config requires a path")
			}
			i++ // Config.Load already applies process overlays.
		case "-dict":
			if i+1 >= len(args) {
				return errors.New("-dict requires a dictionary file")
			}
			i++
			dictionary = args[i]
		case "-json":
			opts.JSON = true
		case "-raw":
			opts.Raw = true
		case "-color":
			opts.Color = true
		case "-f":
			opts.Follow = true
			followSet = true
		case "-cat":
			if selected {
				return errors.New("select only one of -cat, -head and -tail")
			}
			selected = true
		case "-tail", "-head":
			if selected || i+1 >= len(args) {
				return errors.New("head/tail require one positive count")
			}
			i++
			for _, r := range args[i] {
				if r < '0' || r > '9' {
					return errors.New("invalid head/tail count")
				}
			}
			n := cast.Int(args[i])
			if n <= 0 || n > 10000000 {
				return errors.New("head/tail count must be 1–10000000")
			}
			if arg == "-tail" {
				opts.Tail = n
			} else {
				opts.Head = n
			}
			selected = true
		default:
			if strings.HasPrefix(arg, "--config=") || strings.HasPrefix(arg, "-config=") {
				continue
			}
			if strings.HasPrefix(arg, "-") {
				return fmt.Errorf("unknown log option %s", arg)
			}
			paths = append(paths, arg)
		}
	}
	if name != "log" && name != "logv" {
		return errors.New("expected log or logv command")
	}
	if opts.Raw && opts.JSON {
		return errors.New("-raw and -json cannot be combined")
	}
	if name == "logv" && len(paths) == 0 {
		if selected || followSet {
			return errors.New("stdin formatting does not accept range or follow options")
		}
		if dictionary != "" {
			return FormatArchiveStream(ctx, input, output, dictionary, opts)
		}
		return FormatStream(ctx, input, output, opts)
	}
	if dictionary != "" {
		return errors.New("-dict is only valid for compressed logv stdin; files resolve their own dictionary")
	}
	if name == "log" && !selected {
		if len(paths) == 0 || followSet {
			opts.Tail = 100
		}
		if len(paths) == 0 {
			opts.Follow = true
		}
	}
	if opts.Follow && len(paths) > 1 {
		return errors.New("follow accepts only one file")
	}
	if len(paths) == 0 {
		return Query(ctx, DefaultConfig(), output, opts)
	}
	for _, path := range paths {
		if err := QueryFile(ctx, path, output, opts); err != nil {
			return err
		}
	}
	return nil
}
