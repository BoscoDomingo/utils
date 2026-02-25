package main_test

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

var (
	binaryBuildOnce sync.Once
	binaryPath      string
	binaryBuildErr  error
)

func TestCLIFlagsE2E(t *testing.T) {
	t.Parallel()

	runRecursiveCase := func(t *testing.T, recursiveFlag string) {
		t.Helper()

		tempDir := t.TempDir()
		sourceDir := filepath.Join(tempDir, "src")
		if err := os.MkdirAll(filepath.Join(sourceDir, "nested"), 0o755); err != nil {
			t.Fatalf("create source directory: %v", err)
		}

		expectedContents := "recursive-copy"
		sourceFile := filepath.Join(sourceDir, "nested", "payload.txt")
		if err := os.WriteFile(sourceFile, []byte(expectedContents), 0o644); err != nil {
			t.Fatalf("write source file: %v", err)
		}

		destinationDir := filepath.Join(tempDir, "dest")
		stdout, stderr, err := runCLI(t, tempDir, recursiveFlag, sourceDir, destinationDir)
		if err != nil {
			t.Fatalf("recursive copy failed: %v (stdout=%q, stderr=%q)", err, stdout, stderr)
		}
		if stderr != "" {
			t.Fatalf("expected empty stderr, got %q", stderr)
		}
		if !strings.Contains(stdout, "Copied 1 file(s)") {
			t.Fatalf("expected copy summary in stdout, got %q", stdout)
		}

		copiedFile := filepath.Join(destinationDir, "nested", "payload.txt")
		actualBytes, err := os.ReadFile(copiedFile)
		if err != nil {
			t.Fatalf("read copied file: %v", err)
		}
		if string(actualBytes) != expectedContents {
			t.Fatalf("unexpected copied contents: got %q, want %q", actualBytes, expectedContents)
		}
	}

	runForceCase := func(t *testing.T, forceFlag string) {
		t.Helper()

		tempDir := t.TempDir()
		sourceFile := filepath.Join(tempDir, "source.txt")
		destinationFile := filepath.Join(tempDir, "destination.txt")

		if err := os.WriteFile(sourceFile, []byte("new-content"), 0o644); err != nil {
			t.Fatalf("write source file: %v", err)
		}
		if err := os.WriteFile(destinationFile, []byte("old-content"), 0o644); err != nil {
			t.Fatalf("write destination file: %v", err)
		}

		_, stderrWithoutForce, errWithoutForce := runCLI(t, tempDir, sourceFile, destinationFile)
		if errWithoutForce == nil {
			t.Fatalf("expected copy to fail without force when destination exists")
		}
		if !strings.Contains(stderrWithoutForce, "destination file exists") {
			t.Fatalf("expected overwrite error, got stderr=%q", stderrWithoutForce)
		}

		stdout, stderr, err := runCLI(t, tempDir, forceFlag, sourceFile, destinationFile)
		if err != nil {
			t.Fatalf("force copy failed: %v (stdout=%q, stderr=%q)", err, stdout, stderr)
		}
		if stderr != "" {
			t.Fatalf("expected empty stderr, got %q", stderr)
		}

		actualBytes, err := os.ReadFile(destinationFile)
		if err != nil {
			t.Fatalf("read destination file: %v", err)
		}
		if string(actualBytes) != "new-content" {
			t.Fatalf("destination was not overwritten, got %q", string(actualBytes))
		}
	}

	runPreserveCase := func(t *testing.T, preserveFlag string) {
		t.Helper()

		tempDir := t.TempDir()
		sourceFile := filepath.Join(tempDir, "source.txt")
		destinationFile := filepath.Join(tempDir, "destination.txt")

		if err := os.WriteFile(sourceFile, []byte("metadata"), 0o640); err != nil {
			t.Fatalf("write source file: %v", err)
		}

		expectedModTime := time.Now().Add(-3 * time.Hour).Truncate(time.Second)
		if err := os.Chtimes(sourceFile, expectedModTime, expectedModTime); err != nil {
			t.Fatalf("set source times: %v", err)
		}

		stdout, stderr, err := runCLI(t, tempDir, preserveFlag, sourceFile, destinationFile)
		if err != nil {
			t.Fatalf("preserve copy failed: %v (stdout=%q, stderr=%q)", err, stdout, stderr)
		}
		if stderr != "" {
			t.Fatalf("expected empty stderr, got %q", stderr)
		}

		info, err := os.Stat(destinationFile)
		if err != nil {
			t.Fatalf("stat destination file: %v", err)
		}

		if runtime.GOOS != "windows" && info.Mode().Perm() != 0o640 {
			t.Fatalf("expected destination mode 0640, got %o", info.Mode().Perm())
		}

		diff := info.ModTime().Sub(expectedModTime)
		if diff < 0 {
			diff = -diff
		}
		if diff > time.Second {
			t.Fatalf("expected destination modtime near %v, got %v", expectedModTime, info.ModTime())
		}
	}

	runQuietCase := func(t *testing.T, quietFlag string) {
		t.Helper()

		tempDir := t.TempDir()
		sourceFile := filepath.Join(tempDir, "source.bin")
		destinationFile := filepath.Join(tempDir, "destination.bin")

		if err := os.WriteFile(sourceFile, bytes.Repeat([]byte("z"), 128*1024), 0o644); err != nil {
			t.Fatalf("write source file: %v", err)
		}

		stdout, stderr, err := runCLI(t, tempDir, quietFlag, sourceFile, destinationFile)
		if err != nil {
			t.Fatalf("quiet copy failed: %v (stdout=%q, stderr=%q)", err, stdout, stderr)
		}
		if stderr != "" {
			t.Fatalf("expected empty stderr, got %q", stderr)
		}
		if !strings.Contains(stdout, "Copied 1 file(s)") {
			t.Fatalf("expected summary in stdout, got %q", stdout)
		}
		if strings.Contains(stdout, "[") {
			t.Fatalf("expected no progress bar output with quiet flag, got %q", stdout)
		}
	}

	t.Run("recursive_short_flag", func(t *testing.T) {
		t.Parallel()
		runRecursiveCase(t, "-r")
	})

	t.Run("recursive_long_flag", func(t *testing.T) {
		t.Parallel()
		runRecursiveCase(t, "--recursive")
	})

	t.Run("force_short_flag", func(t *testing.T) {
		t.Parallel()
		runForceCase(t, "-f")
	})

	t.Run("force_long_flag", func(t *testing.T) {
		t.Parallel()
		runForceCase(t, "--force")
	})

	t.Run("preserve_short_flag", func(t *testing.T) {
		t.Parallel()
		runPreserveCase(t, "-p")
	})

	t.Run("preserve_long_flag", func(t *testing.T) {
		t.Parallel()
		runPreserveCase(t, "--preserve")
	})

	t.Run("quiet_short_flag", func(t *testing.T) {
		t.Parallel()
		runQuietCase(t, "-q")
	})

	t.Run("quiet_long_flag", func(t *testing.T) {
		t.Parallel()
		runQuietCase(t, "--quiet")
	})

	t.Run("buffer_size_flag", func(t *testing.T) {
		t.Parallel()

		tempDir := t.TempDir()
		sourceFile := filepath.Join(tempDir, "source.bin")
		destinationFile := filepath.Join(tempDir, "destination.bin")

		expectedBytes := bytes.Repeat([]byte("a"), 257*1024)
		if err := os.WriteFile(sourceFile, expectedBytes, 0o644); err != nil {
			t.Fatalf("write source file: %v", err)
		}

		stdout, stderr, err := runCLI(t, tempDir, "--buffer-size", "17", sourceFile, destinationFile)
		if err != nil {
			t.Fatalf("buffer-size copy failed: %v (stdout=%q, stderr=%q)", err, stdout, stderr)
		}
		if stderr != "" {
			t.Fatalf("expected empty stderr, got %q", stderr)
		}

		actualBytes, err := os.ReadFile(destinationFile)
		if err != nil {
			t.Fatalf("read destination file: %v", err)
		}
		if !bytes.Equal(actualBytes, expectedBytes) {
			t.Fatalf("destination content mismatch after buffer-size copy")
		}
	})

	t.Run("buffer_size_validation", func(t *testing.T) {
		t.Parallel()

		tempDir := t.TempDir()
		sourceFile := filepath.Join(tempDir, "source.txt")
		destinationFile := filepath.Join(tempDir, "destination.txt")

		if err := os.WriteFile(sourceFile, []byte("invalid-buffer-size"), 0o644); err != nil {
			t.Fatalf("write source file: %v", err)
		}

		_, stderr, err := runCLI(t, tempDir, "--buffer-size", "0", sourceFile, destinationFile)
		if err == nil {
			t.Fatalf("expected invalid buffer-size to fail")
		}
		if !strings.Contains(stderr, "buffer-size must be greater than 0") {
			t.Fatalf("expected buffer-size validation error, got stderr=%q", stderr)
		}
	})
}

func runCLI(t *testing.T, workingDirectory string, args ...string) (string, string, error) {
	t.Helper()

	command := exec.Command(zcpBinary(t), args...)
	command.Dir = workingDirectory

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr

	err := command.Run()
	return stdout.String(), stderr.String(), err
}

func zcpBinary(t *testing.T) string {
	t.Helper()

	binaryBuildOnce.Do(func() {
		moduleRoot, err := moduleRootPath()
		if err != nil {
			binaryBuildErr = err
			return
		}

		tempDir, err := os.MkdirTemp("", "zcp-e2e-binary-*")
		if err != nil {
			binaryBuildErr = fmt.Errorf("create temp directory for binary: %w", err)
			return
		}

		binaryName := "zcp"
		if runtime.GOOS == "windows" {
			binaryName += ".exe"
		}
		binaryPath = filepath.Join(tempDir, binaryName)

		buildCommand := exec.Command("go", "build", "-o", binaryPath, "./cmd/zcp")
		buildCommand.Dir = moduleRoot

		buildOutput, err := buildCommand.CombinedOutput()
		if err != nil {
			binaryBuildErr = fmt.Errorf("build zcp binary: %w\n%s", err, string(buildOutput))
		}
	})

	if binaryBuildErr != nil {
		t.Fatalf("prepare zcp binary: %v", binaryBuildErr)
	}

	return binaryPath
}

func moduleRootPath() (string, error) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("resolve current file for module path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", "..")), nil
}
