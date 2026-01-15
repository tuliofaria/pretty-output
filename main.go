package main

import (
	"fmt"
	"os"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"

	"pretty-output/parser"
	"pretty-output/store"
	"pretty-output/ui"
)

func main() {
	// Check if stdin is a pipe
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		fmt.Fprintln(os.Stderr, "Usage: docker compose up | pretty-output")
		fmt.Fprintln(os.Stderr, "       docker logs -f container | pretty-output")
		os.Exit(1)
	}

	// Create store
	logStore := store.New()

	// Create TUI program
	model := ui.New(logStore)
	p := tea.NewProgram(model, tea.WithAltScreen())

	// Channel for parsed entries
	entries := make(chan parser.Entry, 100)

	// Start parser in goroutine
	go parser.Parse(os.Stdin, entries)

	// Forward entries to TUI
	go func() {
		for entry := range entries {
			p.Send(ui.NewEntryMsg{Entry: entry})
		}
	}()

	// Run TUI
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Send SIGINT to the process group to stop docker compose
	// Kill(0, ...) sends to all processes in current process group
	syscall.Kill(0, syscall.SIGINT)
}
