package zcp

import (
	"fmt"
	"io"
	"math"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type progressBar struct {
	total      uint64
	startedAt  time.Time
	completed  atomic.Uint64
	enabled    bool
	writer     io.Writer
	terminal   bool
	stopCh     chan struct{}
	stopOnce   sync.Once
	waitGroup  sync.WaitGroup
	lastRender int
}

func newProgressBar(total uint64, enabled bool, writer io.Writer) *progressBar {
	bar := &progressBar{
		total:    total,
		enabled:  enabled && total > 0,
		writer:   writer,
		terminal: isTerminalWriter(writer),
	}
	return bar
}

func (p *progressBar) start() {
	if !p.enabled {
		return
	}

	p.stopCh = make(chan struct{})
	p.startedAt = time.Now()
	p.waitGroup.Add(1)

	go func() {
		defer p.waitGroup.Done()

		interval := 120 * time.Millisecond
		if !p.terminal {
			interval = time.Second
		}

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				p.render(false)
			case <-p.stopCh:
				p.render(true)
				return
			}
		}
	}()
}

func (p *progressBar) stop() {
	if !p.enabled {
		return
	}

	p.stopOnce.Do(func() {
		close(p.stopCh)
		p.waitGroup.Wait()
	})
}

func (p *progressBar) add(value uint64) {
	if !p.enabled || value == 0 {
		return
	}
	p.completed.Add(value)
}

func (p *progressBar) render(final bool) {
	done := p.completed.Load()
	if done > p.total {
		done = p.total
	}
	if final {
		done = p.total
	}

	elapsed := time.Since(p.startedAt)
	if elapsed <= 0 {
		elapsed = time.Millisecond
	}

	bytesPerSecond := float64(done) / elapsed.Seconds()
	line := formatProgressLine(done, p.total, bytesPerSecond)

	if p.terminal {
		padding := ""
		if len(line) < p.lastRender {
			padding = strings.Repeat(" ", p.lastRender-len(line))
		}
		fmt.Fprintf(p.writer, "\r%s%s", line, padding)
		p.lastRender = len(line)
		if final {
			fmt.Fprint(p.writer, "\n")
		}
		return
	}

	fmt.Fprintln(p.writer, line)
}

func formatProgressLine(done uint64, total uint64, bytesPerSecond float64) string {
	if total == 0 {
		return "[==============================] 100.00% 0 B/0 B 0 B/s ETA 00:00"
	}

	percentage := float64(done) / float64(total) * 100
	if percentage > 100 {
		percentage = 100
	}

	eta := "00:00"
	if done < total && bytesPerSecond > 0 {
		remainingSeconds := float64(total-done) / bytesPerSecond
		eta = formatDuration(time.Duration(remainingSeconds * float64(time.Second)))
	}

	return fmt.Sprintf(
		"[%s] %6.2f%% %s/%s %s/s ETA %s",
		buildBar(percentage, 30),
		percentage,
		humanizeBytes(done),
		humanizeBytes(total),
		humanizeRate(bytesPerSecond),
		eta,
	)
}

func buildBar(percentage float64, width int) string {
	if width <= 0 {
		return ""
	}

	filled := int(math.Round((percentage / 100) * float64(width)))
	if filled < 0 {
		filled = 0
	}
	if filled > width {
		filled = width
	}

	switch {
	case filled <= 0:
		return strings.Repeat(" ", width)
	case filled >= width:
		return strings.Repeat("=", width)
	default:
		return strings.Repeat("=", filled-1) + ">" + strings.Repeat(" ", width-filled)
	}
}

func humanizeBytes(value uint64) string {
	const unit = 1024.0
	if value < 1024 {
		return fmt.Sprintf("%d B", value)
	}

	size := float64(value)
	units := []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB"}
	unitIndex := 0
	for size >= unit && unitIndex < len(units)-1 {
		size /= unit
		unitIndex++
	}

	return fmt.Sprintf("%.1f %s", size, units[unitIndex])
}

func humanizeRate(bytesPerSecond float64) string {
	if bytesPerSecond <= 0 {
		return "0 B"
	}

	units := []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB"}
	value := bytesPerSecond
	unitIndex := 0
	for value >= 1024 && unitIndex < len(units)-1 {
		value /= 1024
		unitIndex++
	}

	if unitIndex == 0 {
		return fmt.Sprintf("%.0f %s", value, units[unitIndex])
	}
	return fmt.Sprintf("%.1f %s", value, units[unitIndex])
}

func formatDuration(duration time.Duration) string {
	if duration < 0 {
		duration = 0
	}

	totalSeconds := int64(duration.Round(time.Second).Seconds())
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60

	if hours > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
	}
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

func isTerminalWriter(writer io.Writer) bool {
	file, ok := writer.(*os.File)
	if !ok {
		return false
	}

	info, err := file.Stat()
	if err != nil {
		return false
	}

	return info.Mode()&os.ModeCharDevice != 0
}
