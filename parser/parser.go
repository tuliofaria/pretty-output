package parser

import (
	"bufio"
	"encoding/json"
	"io"
	"regexp"
	"strings"
)

// Entry represents a parsed log line
type Entry struct {
	Container string
	Content   string
	IsJSON    bool
	JSON      map[string]any
}

// dockerComposePattern matches lines like "container-name  | content"
var dockerComposePattern = regexp.MustCompile(`^([a-zA-Z0-9_-]+(?:-\d+)?)\s+\|\s+(.*)$`)

// ansiPattern matches ANSI escape codes
var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

// Parse reads from the given reader and sends parsed entries to the channel.
// It closes the channel when done.
func Parse(r io.Reader, entries chan<- Entry) {
	defer close(entries)

	scanner := bufio.NewScanner(r)
	// Increase buffer size for long JSON lines
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		entry := parseLine(line)
		entries <- entry
	}
}

func parseLine(line string) Entry {
	// Strip ANSI escape codes
	line = ansiPattern.ReplaceAllString(line, "")

	entry := Entry{
		Container: "default",
		Content:   line,
	}

	// Try to match docker compose format
	if matches := dockerComposePattern.FindStringSubmatch(line); matches != nil {
		entry.Container = matches[1]
		entry.Content = matches[2]
	}

	// Try to parse as JSON
	content := strings.TrimSpace(entry.Content)
	if len(content) > 0 && (content[0] == '{' || content[0] == '[') {
		var parsed map[string]any
		if err := json.Unmarshal([]byte(content), &parsed); err == nil {
			entry.IsJSON = true
			entry.JSON = parsed
		}
	}

	return entry
}
