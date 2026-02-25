# zcp

Like `cp`, but with a progress bar. The `z` is there because I liked it and to avoid name collisions.

Inspired by the original `gcp` utility:
https://manpages.ubuntu.com/manpages/focal/man1/gcp.1.html

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
- Optional verbose output (`-v`) to print created file names

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

Verbose file listing:

```bash
zcp -v -r photos /mnt/backup/
```

## Options

- `-r`, `--recursive`: copy directories recursively
- `-f`, `--force`: overwrite destination files
- `-p`, `--preserve`: preserve mode and modification time
- `-q`, `--quiet`: disable progress output
- `-v`, `--verbose`: print created file names
- `--buffer-size`: copy buffer size in bytes (default `1048576`)

## Build

From this directory:

```bash
mkdir -p bin
go build -o bin/zcp ./cmd/zcp
```

### Cross-compile

Linux:

```bash
mkdir -p bin
GOOS=linux GOARCH=amd64 go build -o bin/zcp-linux-amd64 ./cmd/zcp
```

Windows:

```bash
mkdir -p bin
GOOS=windows GOARCH=amd64 go build -o bin/zcp-windows-amd64.exe ./cmd/zcp
```

## Notes

- Symbolic links are currently not copied.
- For multiple sources, destination must already exist as a directory.
