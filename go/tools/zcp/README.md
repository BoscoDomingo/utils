# zcp (Zippy Copy with Progress)

`zcp` is a small cross-platform CLI (Linux/Windows/macOS) that copies files and directories while showing a live progress bar.

It is inspired by the Python `gcp` utility but implemented in pure Go.

## Features

- Copy files and directories
- Recursive directory copy (`-r`)
- Per-copy progress bar with:
  - percent complete
  - bytes copied / total bytes
  - transfer speed
  - ETA
- Optional metadata preservation (`-p` mode + mtime)
- Optional overwrite (`-f`)

## Usage

```bash
zcp [options] SOURCE... DEST
```

### Examples

Copy a single file:

```bash
zcp movie.mkv /mnt/backup/movie.mkv
```

Copy a directory recursively:

```bash
zcp -r photos /mnt/backup/
```

Copy multiple sources into an existing destination directory:

```bash
zcp -r folder_a folder_b file.txt /mnt/backup/
```

Overwrite existing files:

```bash
zcp -f large.iso /mnt/backup/large.iso
```

Preserve source mode + mtime:

```bash
zcp -p -r assets ./assets-copy
```

Disable progress output:

```bash
zcp -q -r logs /tmp/logs-copy
```

## Options

- `-r`, `--recursive`: copy directories recursively
- `-f`, `--force`: overwrite destination files
- `-p`, `--preserve`: preserve mode and modification time
- `-q`, `--quiet`: disable progress output
- `--buffer-size`: copy buffer size in bytes (default `1048576`)

## Build

From this directory:

```bash
go build -o bin/zcp .
```

### Cross-compile

Linux:

```bash
GOOS=linux GOARCH=amd64 go build -o bin/zcp-linux-amd64 .
```

Windows:

```bash
GOOS=windows GOARCH=amd64 go build -o bin/zcp-windows-amd64.exe .
```

## Notes

- Symbolic links are currently not copied.
- For multiple sources, destination must already exist as a directory.
