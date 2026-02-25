package zcp

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type operationType int

const (
	operationCreateDirectory operationType = iota
	operationCopyFile
)

type copyOperation struct {
	kind        operationType
	source      string
	destination string
	mode        fs.FileMode
	modTime     time.Time
	size        uint64
}

func buildCopyPlan(sources []string, destination string, recursive bool) ([]copyOperation, uint64, error) {
	destInfo, destErr := os.Stat(destination)
	destExists := destErr == nil
	if destErr != nil && !errors.Is(destErr, os.ErrNotExist) {
		return nil, 0, fmt.Errorf("stat destination %q: %w", destination, destErr)
	}

	destIsDir := destExists && destInfo.IsDir()
	if len(sources) > 1 && !destIsDir {
		return nil, 0, fmt.Errorf("destination %q must be an existing directory when copying multiple sources", destination)
	}

	plan := make([]copyOperation, 0, len(sources))
	var totalBytes uint64

	for _, source := range sources {
		source = filepath.Clean(source)

		sourceInfo, err := os.Lstat(source)
		if err != nil {
			return nil, 0, fmt.Errorf("stat source %q: %w", source, err)
		}

		if sourceInfo.Mode()&os.ModeSymlink != 0 {
			return nil, 0, fmt.Errorf("symbolic links are not supported: %q", source)
		}

		target := destination
		if len(sources) > 1 || destIsDir {
			target = filepath.Join(destination, filepath.Base(source))
		}

		if sourceInfo.IsDir() {
			if !recursive {
				return nil, 0, fmt.Errorf("omitting directory %q (use -r or --recursive)", source)
			}

			if destExists && !destIsDir && len(sources) == 1 {
				return nil, 0, fmt.Errorf("cannot overwrite non-directory %q with directory %q", destination, source)
			}

			if err := ensureDestinationOutsideSource(source, target); err != nil {
				return nil, 0, err
			}

			directoryOps, directoryBytes, err := collectDirectoryOperations(source, target)
			if err != nil {
				return nil, 0, err
			}

			plan = append(plan, directoryOps...)
			totalBytes += directoryBytes
			continue
		}

		sameFile, err := refersToSameFile(source, target)
		if err != nil {
			return nil, 0, err
		}
		if sameFile {
			return nil, 0, fmt.Errorf("%q and %q are the same file", source, target)
		}

		size := sourceInfo.Size()
		if size < 0 {
			size = 0
		}

		plan = append(plan, copyOperation{
			kind:        operationCopyFile,
			source:      source,
			destination: target,
			mode:        sourceInfo.Mode(),
			modTime:     sourceInfo.ModTime(),
			size:        uint64(size),
		})
		totalBytes += uint64(size)
	}

	return plan, totalBytes, nil
}

func collectDirectoryOperations(sourceRoot string, destinationRoot string) ([]copyOperation, uint64, error) {
	operations := make([]copyOperation, 0, 16)
	var totalBytes uint64

	err := filepath.WalkDir(sourceRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("symbolic links are not supported: %q", path)
		}

		entryInfo, err := entry.Info()
		if err != nil {
			return err
		}

		relativePath, err := filepath.Rel(sourceRoot, path)
		if err != nil {
			return err
		}

		destinationPath := destinationRoot
		if relativePath != "." {
			destinationPath = filepath.Join(destinationRoot, relativePath)
		}

		if entry.IsDir() {
			operations = append(operations, copyOperation{
				kind:        operationCreateDirectory,
				source:      path,
				destination: destinationPath,
				mode:        entryInfo.Mode(),
				modTime:     entryInfo.ModTime(),
			})
			return nil
		}

		size := entryInfo.Size()
		if size < 0 {
			size = 0
		}

		operations = append(operations, copyOperation{
			kind:        operationCopyFile,
			source:      path,
			destination: destinationPath,
			mode:        entryInfo.Mode(),
			modTime:     entryInfo.ModTime(),
			size:        uint64(size),
		})
		totalBytes += uint64(size)
		return nil
	})
	if err != nil {
		return nil, 0, fmt.Errorf("walk source directory %q: %w", sourceRoot, err)
	}

	return operations, totalBytes, nil
}

func ensureDestinationOutsideSource(sourceDirectory string, destinationPath string) error {
	sourceAbs, err := filepath.Abs(sourceDirectory)
	if err != nil {
		return fmt.Errorf("resolve source path %q: %w", sourceDirectory, err)
	}

	destinationAbs, err := filepath.Abs(destinationPath)
	if err != nil {
		return fmt.Errorf("resolve destination path %q: %w", destinationPath, err)
	}

	relative, err := filepath.Rel(sourceAbs, destinationAbs)
	if err != nil {
		return fmt.Errorf("check destination relation: %w", err)
	}

	if relative == "." || relative == "" {
		return fmt.Errorf("cannot copy %q to itself", sourceDirectory)
	}

	parentPrefix := ".." + string(os.PathSeparator)
	if relative == ".." || strings.HasPrefix(relative, parentPrefix) {
		return nil
	}

	return fmt.Errorf("cannot copy directory %q into itself (%q)", sourceDirectory, destinationPath)
}

func refersToSameFile(source string, destination string) (bool, error) {
	sourceInfo, err := os.Stat(source)
	if err != nil {
		return false, fmt.Errorf("stat source %q: %w", source, err)
	}

	destinationInfo, err := os.Stat(destination)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, fmt.Errorf("stat destination %q: %w", destination, err)
	}

	return os.SameFile(sourceInfo, destinationInfo), nil
}

func executePlan(plan []copyOperation, opts options, progress *progressBar) error {
	directoriesToPreserve := make([]copyOperation, 0)

	for _, op := range plan {
		switch op.kind {
		case operationCreateDirectory:
			if err := os.MkdirAll(op.destination, op.mode.Perm()); err != nil {
				return fmt.Errorf("create directory %q: %w", op.destination, err)
			}
			if opts.preserve {
				directoriesToPreserve = append(directoriesToPreserve, op)
			}

		case operationCopyFile:
			if err := copyFile(op, opts, progress); err != nil {
				return err
			}

		default:
			return fmt.Errorf("unsupported copy operation: %v", op.kind)
		}
	}

	if opts.preserve {
		for i := len(directoriesToPreserve) - 1; i >= 0; i-- {
			directory := directoriesToPreserve[i]
			if err := setMetadata(directory.destination, directory.mode, directory.modTime); err != nil {
				return err
			}
		}
	}

	return nil
}

func copyFile(op copyOperation, opts options, progress *progressBar) error {
	if err := os.MkdirAll(filepath.Dir(op.destination), 0o755); err != nil {
		return fmt.Errorf("create destination parent for %q: %w", op.destination, err)
	}

	sourceFile, err := os.Open(op.source)
	if err != nil {
		return fmt.Errorf("open source file %q: %w", op.source, err)
	}
	defer sourceFile.Close()

	flags := os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	if !opts.force {
		flags |= os.O_EXCL
	}

	destinationFile, err := os.OpenFile(op.destination, flags, op.mode.Perm())
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return fmt.Errorf("destination file exists (use -f to overwrite): %q", op.destination)
		}
		return fmt.Errorf("open destination file %q: %w", op.destination, err)
	}

	buffer := make([]byte, opts.bufferSize)
	for {
		readBytes, readErr := sourceFile.Read(buffer)
		if readBytes > 0 {
			writtenBytes, writeErr := destinationFile.Write(buffer[:readBytes])
			if writeErr != nil {
				destinationFile.Close()
				return fmt.Errorf("write destination file %q: %w", op.destination, writeErr)
			}
			if writtenBytes != readBytes {
				destinationFile.Close()
				return fmt.Errorf("write destination file %q: short write", op.destination)
			}
			progress.add(uint64(writtenBytes))
		}

		if errors.Is(readErr, io.EOF) {
			break
		}
		if readErr != nil {
			destinationFile.Close()
			return fmt.Errorf("read source file %q: %w", op.source, readErr)
		}
	}

	if err := destinationFile.Close(); err != nil {
		return fmt.Errorf("close destination file %q: %w", op.destination, err)
	}

	if opts.preserve {
		if err := setMetadata(op.destination, op.mode, op.modTime); err != nil {
			return err
		}
	}

	return nil
}

func setMetadata(path string, mode fs.FileMode, modTime time.Time) error {
	if err := os.Chmod(path, mode.Perm()); err != nil {
		return fmt.Errorf("set mode on %q: %w", path, err)
	}
	if err := os.Chtimes(path, modTime, modTime); err != nil {
		return fmt.Errorf("set modification time on %q: %w", path, err)
	}
	return nil
}
