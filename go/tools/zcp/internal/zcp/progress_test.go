package zcp

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestProgressFormattingHelpers(t *testing.T) {
	t.Parallel()

	t.Run("build_bar", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			name       string
			percentage float64
			width      int
			want       string
		}{
			{
				name:       "zero_percent",
				percentage: 0,
				width:      10,
				want:       "          ",
			},
			{
				name:       "half_percent",
				percentage: 50,
				width:      10,
				want:       "====>     ",
			},
			{
				name:       "full_percent",
				percentage: 100,
				width:      10,
				want:       "==========",
			},
			{
				name:       "clamps_above_hundred",
				percentage: 150,
				width:      10,
				want:       "==========",
			},
			{
				name:       "zero_width",
				percentage: 50,
				width:      0,
				want:       "",
			},
		}

		for _, testCase := range testCases {
			testCase := testCase
			t.Run(testCase.name, func(t *testing.T) {
				t.Parallel()

				got := buildBar(testCase.percentage, testCase.width)
				if got != testCase.want {
					t.Fatalf("buildBar(%v, %d) = %q, want %q", testCase.percentage, testCase.width, got, testCase.want)
				}
			})
		}
	})

	t.Run("humanize_bytes", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			name  string
			value uint64
			want  string
		}{
			{name: "bytes", value: 512, want: "512 B"},
			{name: "kib", value: 1024, want: "1.0 KiB"},
			{name: "fractional_kib", value: 1536, want: "1.5 KiB"},
			{name: "mib", value: 1024 * 1024, want: "1.0 MiB"},
		}

		for _, testCase := range testCases {
			testCase := testCase
			t.Run(testCase.name, func(t *testing.T) {
				t.Parallel()

				got := humanizeBytes(testCase.value)
				if got != testCase.want {
					t.Fatalf("humanizeBytes(%d) = %q, want %q", testCase.value, got, testCase.want)
				}
			})
		}
	})

	t.Run("humanize_rate", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			name  string
			value float64
			want  string
		}{
			{name: "zero", value: 0, want: "0 B"},
			{name: "bytes", value: 256, want: "256 B"},
			{name: "kib", value: 1024, want: "1.0 KiB"},
			{name: "fractional_kib", value: 1536, want: "1.5 KiB"},
		}

		for _, testCase := range testCases {
			testCase := testCase
			t.Run(testCase.name, func(t *testing.T) {
				t.Parallel()

				got := humanizeRate(testCase.value)
				if got != testCase.want {
					t.Fatalf("humanizeRate(%f) = %q, want %q", testCase.value, got, testCase.want)
				}
			})
		}
	})

	t.Run("format_duration", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			name  string
			value time.Duration
			want  string
		}{
			{name: "seconds", value: 59 * time.Second, want: "00:59"},
			{name: "minutes", value: 2*time.Minute + 5*time.Second, want: "02:05"},
			{name: "hours", value: time.Hour + 2*time.Minute + 3*time.Second, want: "01:02:03"},
		}

		for _, testCase := range testCases {
			testCase := testCase
			t.Run(testCase.name, func(t *testing.T) {
				t.Parallel()

				got := formatDuration(testCase.value)
				if got != testCase.want {
					t.Fatalf("formatDuration(%v) = %q, want %q", testCase.value, got, testCase.want)
				}
			})
		}
	})

	t.Run("format_progress_line", func(t *testing.T) {
		t.Parallel()

		line := formatProgressLine(512, 1024, 256)
		expectedFragments := []string{
			"50.00%",
			"512 B/1.0 KiB",
			"256 B/s",
			"ETA 00:02",
		}
		for _, fragment := range expectedFragments {
			if !strings.Contains(line, fragment) {
				t.Fatalf("expected fragment %q in line %q", fragment, line)
			}
		}

		zeroTotalLine := formatProgressLine(0, 0, 0)
		if !strings.Contains(zeroTotalLine, "100.00% 0 B/0 B") {
			t.Fatalf("unexpected zero-total line: %q", zeroTotalLine)
		}
	})
}

func TestProgressBarLifecycle(t *testing.T) {
	t.Parallel()

	t.Run("writes_final_line_on_stop", func(t *testing.T) {
		t.Parallel()

		var output bytes.Buffer
		bar := newProgressBar(1024, true, &output)
		bar.start()
		bar.add(1024)
		bar.stop()

		got := output.String()
		if !strings.Contains(got, "100.00%") {
			t.Fatalf("expected final progress output, got %q", got)
		}
	})

	t.Run("disabled_for_zero_total", func(t *testing.T) {
		t.Parallel()

		var output bytes.Buffer
		bar := newProgressBar(0, true, &output)
		bar.start()
		bar.add(100)
		bar.stop()

		if output.Len() != 0 {
			t.Fatalf("expected no output for zero-total progress, got %q", output.String())
		}
	})
}
