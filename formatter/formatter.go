package formatter

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"pretty-output/parser"
)

var (
	keyStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("86"))  // Cyan
	stringStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("82"))  // Green
	numberStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("220")) // Yellow
	boolStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("213")) // Magenta
	nullStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("245")) // Gray
	bracketStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("250")) // Light gray
	nonJSONStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245")) // Dim gray
)

// Format formats a log entry with syntax highlighting
func Format(entry parser.Entry) string {
	if !entry.IsJSON {
		return nonJSONStyle.Render(entry.Content)
	}
	return formatValue(entry.JSON, 0)
}

func formatValue(v any, indent int) string {
	switch val := v.(type) {
	case map[string]any:
		return formatObject(val, indent)
	case []any:
		return formatArray(val, indent)
	case string:
		return stringStyle.Render(fmt.Sprintf("%q", val))
	case float64:
		// Check if it's an integer
		if val == float64(int64(val)) {
			return numberStyle.Render(fmt.Sprintf("%.0f", val))
		}
		return numberStyle.Render(fmt.Sprintf("%g", val))
	case bool:
		return boolStyle.Render(fmt.Sprintf("%v", val))
	case nil:
		return nullStyle.Render("null")
	default:
		return fmt.Sprintf("%v", val)
	}
}

func formatObject(obj map[string]any, indent int) string {
	if len(obj) == 0 {
		return bracketStyle.Render("{}")
	}

	var sb strings.Builder
	sb.WriteString(bracketStyle.Render("{"))
	sb.WriteString("\n")

	// Sort keys for consistent output
	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	indentStr := strings.Repeat("  ", indent+1)
	for i, key := range keys {
		sb.WriteString(indentStr)
		sb.WriteString(keyStyle.Render(fmt.Sprintf("%q", key)))
		sb.WriteString(": ")
		sb.WriteString(formatValue(obj[key], indent+1))
		if i < len(keys)-1 {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}

	sb.WriteString(strings.Repeat("  ", indent))
	sb.WriteString(bracketStyle.Render("}"))
	return sb.String()
}

func formatArray(arr []any, indent int) string {
	if len(arr) == 0 {
		return bracketStyle.Render("[]")
	}

	// Check if array contains only primitives and is short
	allPrimitive := true
	for _, v := range arr {
		switch v.(type) {
		case map[string]any, []any:
			allPrimitive = false
		}
	}

	if allPrimitive && len(arr) <= 5 {
		// Inline short primitive arrays
		var parts []string
		for _, v := range arr {
			parts = append(parts, formatValue(v, indent))
		}
		return bracketStyle.Render("[") + strings.Join(parts, ", ") + bracketStyle.Render("]")
	}

	var sb strings.Builder
	sb.WriteString(bracketStyle.Render("["))
	sb.WriteString("\n")

	indentStr := strings.Repeat("  ", indent+1)
	for i, v := range arr {
		sb.WriteString(indentStr)
		sb.WriteString(formatValue(v, indent+1))
		if i < len(arr)-1 {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}

	sb.WriteString(strings.Repeat("  ", indent))
	sb.WriteString(bracketStyle.Render("]"))
	return sb.String()
}
