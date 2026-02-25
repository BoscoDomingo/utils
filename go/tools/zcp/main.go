package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
)

const defaultBufferSize = 1024 * 1024

type options struct {
	recursive  bool
	force      bool
	preserve   bool
	quiet      bool
	bufferSize int
}

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "zcp: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer, stderr io.Writer) error {
	opts, sources, destination, err := parseArgs(args, stderr)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	plan, totalBytes, err := buildCopyPlan(sources, destination, opts.recursive)
	if err != nil {
		return err
	}

	progress := newProgressBar(totalBytes, !opts.quiet, stdout)
	progress.start()
	defer progress.stop()

	if err := executePlan(plan, opts, progress); err != nil {
		return err
	}
	progress.stop()

	fmt.Fprintf(stdout, "Copied %d file(s), %s total.\n", countFiles(plan), humanizeBytes(totalBytes))
	return nil
}

func parseArgs(args []string, stderr io.Writer) (options, []string, string, error) {
	opts := options{
		bufferSize: defaultBufferSize,
	}

	fs := flag.NewFlagSet("zcp", flag.ContinueOnError)
	fs.SetOutput(stderr)

	fs.BoolVar(&opts.recursive, "r", false, "copy directories recursively")
	fs.BoolVar(&opts.recursive, "recursive", false, "copy directories recursively")
	fs.BoolVar(&opts.force, "f", false, "overwrite destination files if they already exist")
	fs.BoolVar(&opts.force, "force", false, "overwrite destination files if they already exist")
	fs.BoolVar(&opts.preserve, "p", false, "preserve file mode and modification time")
	fs.BoolVar(&opts.preserve, "preserve", false, "preserve file mode and modification time")
	fs.BoolVar(&opts.quiet, "q", false, "disable progress output")
	fs.BoolVar(&opts.quiet, "quiet", false, "disable progress output")
	fs.IntVar(&opts.bufferSize, "buffer-size", defaultBufferSize, "copy buffer size in bytes")

	fs.Usage = func() {
		fmt.Fprintln(stderr, "zcp: copy files and directories with a progress bar")
		fmt.Fprintln(stderr)
		fmt.Fprintln(stderr, "Usage:")
		fmt.Fprintln(stderr, "  zcp [options] SOURCE... DEST")
		fmt.Fprintln(stderr)
		fmt.Fprintln(stderr, "Options:")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return options{}, nil, "", err
	}

	if opts.bufferSize <= 0 {
		return options{}, nil, "", fmt.Errorf("buffer-size must be greater than 0")
	}

	remaining := fs.Args()
	if len(remaining) < 2 {
		fs.Usage()
		return options{}, nil, "", fmt.Errorf("expected at least one SOURCE and one DEST")
	}

	return opts, remaining[:len(remaining)-1], remaining[len(remaining)-1], nil
}

func countFiles(plan []copyOperation) int {
	count := 0
	for _, op := range plan {
		if op.kind == operationCopyFile {
			count++
		}
	}
	return count
}
