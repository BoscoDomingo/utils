package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestBuildCopyPlanRequiresRecursiveForDirectories(t *testing.T) {
	tempDir := t.TempDir()
	sourceDir := filepath.Join(tempDir, "source")
	if err := os.MkdirAll(sourceDir, 0o755); err != nil {
		t.Fatalf("mkdir source: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "file.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatalf("write source file: %v", err)
	}

	_, _, err := buildCopyPlan([]string{sourceDir}, filepath.Join(tempDir, "dest"), false)
	if err == nil {
		t.Fatalf("expected error for missing recursive flag")
	}
	if !strings.Contains(err.Error(), "use -r") {
		t.Fatalf("expected recursive hint, got: %v", err)
	}
}

func TestBuildCopyPlanRequiresDirectoryForMultipleSources(t *testing.T) {
	tempDir := t.TempDir()
	first := filepath.Join(tempDir, "first.txt")
	second := filepath.Join(tempDir, "second.txt")
	if err := os.WriteFile(first, []byte("a"), 0o644); err != nil {
		t.Fatalf("write first source: %v", err)
	}
	if err := os.WriteFile(second, []byte("b"), 0o644); err != nil {
		t.Fatalf("write second source: %v", err)
	}

	notDirectory := filepath.Join(tempDir, "dest.txt")
	if err := os.WriteFile(notDirectory, []byte("existing"), 0o644); err != nil {
		t.Fatalf("write destination file: %v", err)
	}

	_, _, err := buildCopyPlan([]string{first, second}, notDirectory, false)
	if err == nil {
		t.Fatalf("expected error for multiple sources to non-directory destination")
	}
}

func TestExecutePlanCopiesNestedDirectory(t *testing.T) {
	tempDir := t.TempDir()
	sourceRoot := filepath.Join(tempDir, "source")
	nestedDir := filepath.Join(sourceRoot, "nested")
	if err := os.MkdirAll(nestedDir, 0o755); err != nil {
		t.Fatalf("mkdir nested: %v", err)
	}

	sourceFile := filepath.Join(nestedDir, "payload.txt")
	expectedContents := "copy me"
	if err := os.WriteFile(sourceFile, []byte(expectedContents), 0o640); err != nil {
		t.Fatalf("write source file: %v", err)
	}

	originalModTime := time.Now().Add(-2 * time.Hour).Truncate(time.Second)
	if err := os.Chtimes(sourceFile, originalModTime, originalModTime); err != nil {
		t.Fatalf("set source modtime: %v", err)
	}

	destinationRoot := filepath.Join(tempDir, "destination")
	plan, totalBytes, err := buildCopyPlan([]string{sourceRoot}, destinationRoot, true)
	if err != nil {
		t.Fatalf("build copy plan: %v", err)
	}
	if totalBytes == 0 {
		t.Fatalf("expected non-zero total bytes")
	}

	opts := options{
		recursive:  true,
		force:      false,
		preserve:   true,
		quiet:      true,
		bufferSize: 8,
	}
	progress := newProgressBar(totalBytes, false, io.Discard)
	if err := executePlan(plan, opts, progress); err != nil {
		t.Fatalf("execute plan: %v", err)
	}

	destinationFile := filepath.Join(destinationRoot, "nested", "payload.txt")
	actualBytes, err := os.ReadFile(destinationFile)
	if err != nil {
		t.Fatalf("read destination file: %v", err)
	}
	if string(actualBytes) != expectedContents {
		t.Fatalf("unexpected destination content: %q", actualBytes)
	}

	info, err := os.Stat(destinationFile)
	if err != nil {
		t.Fatalf("stat destination file: %v", err)
	}

	if !info.ModTime().Equal(originalModTime) {
		t.Fatalf("expected modtime %v, got %v", originalModTime, info.ModTime())
	}
}

func TestExecutePlanHonorsForceFlag(t *testing.T) {
	tempDir := t.TempDir()
	sourceFile := filepath.Join(tempDir, "source.txt")
	destinationFile := filepath.Join(tempDir, "dest.txt")

	if err := os.WriteFile(sourceFile, []byte("new"), 0o644); err != nil {
		t.Fatalf("write source file: %v", err)
	}
	if err := os.WriteFile(destinationFile, []byte("old"), 0o644); err != nil {
		t.Fatalf("write destination file: %v", err)
	}

	plan, totalBytes, err := buildCopyPlan([]string{sourceFile}, destinationFile, false)
	if err != nil {
		t.Fatalf("build copy plan: %v", err)
	}

	noForce := options{
		force:      false,
		bufferSize: 4,
	}
	if err := executePlan(plan, noForce, newProgressBar(totalBytes, false, io.Discard)); err == nil {
		t.Fatalf("expected overwrite error without -f")
	}

	withForce := options{
		force:      true,
		bufferSize: 4,
	}
	if err := executePlan(plan, withForce, newProgressBar(totalBytes, false, io.Discard)); err != nil {
		t.Fatalf("force overwrite failed: %v", err)
	}

	actual, err := os.ReadFile(destinationFile)
	if err != nil {
		t.Fatalf("read destination file: %v", err)
	}
	if string(actual) != "new" {
		t.Fatalf("expected destination content to be overwritten, got %q", string(actual))
	}
}
