# pretty-output

A terminal UI for viewing Docker Compose and container logs with JSON syntax highlighting.

![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go&logoColor=white)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

## Features

- **Container-organized logs** - Automatically groups logs by container name
- **JSON pretty-printing** - Detects and formats JSON with syntax highlighting
- **Real-time streaming** - Updates as new logs arrive
- **Keyboard navigation** - Vim-style keybindings for efficient browsing
- **Log filtering** - Search through logs with `/`

## Installation

```bash
go install github.com/tuliofaria/pretty-output@latest
```

Or build from source:

```bash
git clone https://github.com/tuliofaria/pretty-output
cd pretty-output
go build
```

## Usage

Pipe Docker output to `pretty-output`:

```bash
# Docker Compose logs (stdout and stderr)
docker compose up 2>&1 | pretty-output

# Follow container logs
docker logs -f my-container 2>&1 | pretty-output

# Multiple containers
docker compose logs -f 2>&1 | pretty-output
```

## Keybindings

### Container List (left panel)

| Key                 | Action                    |
| ------------------- | ------------------------- |
| `↑` / `k`           | Select previous container |
| `↓` / `j`           | Select next container     |
| `Enter` / `→` / `l` | View logs                 |
| `Tab`               | Switch to logs panel      |

### Log View (right panel)

| Key               | Action                 |
| ----------------- | ---------------------- |
| `↑` / `k`         | Scroll up              |
| `↓` / `j`         | Scroll down            |
| `PgUp` / `PgDn`   | Scroll half page       |
| `g` / `Home`      | Go to top              |
| `G` / `End`       | Go to bottom           |
| `/`               | Enter filter mode      |
| `←` / `h` / `Esc` | Back to container list |

### Filter Mode

| Key     | Action                |
| ------- | --------------------- |
| Type    | Add to filter         |
| `Enter` | Confirm filter        |
| `Esc`   | Clear filter and exit |

### Global

| Key            | Action |
| -------------- | ------ |
| `q` / `Ctrl+C` | Quit   |

## How It Works

`pretty-output` reads from stdin and parses each line looking for:

1. **Docker Compose format**: Lines matching `container-name  | content`
2. **JSON content**: Automatically detected and pretty-printed with colors

Logs are organized by container in the left panel, with the selected container's logs displayed in the right panel.

## License

[MIT](LICENSE)
