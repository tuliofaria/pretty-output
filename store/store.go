package store

import (
	"sort"
	"sync"

	"pretty-output/parser"
)

// Store holds log entries organized by container
type Store struct {
	mu         sync.RWMutex
	entries    map[string][]parser.Entry
	containers []string
}

// New creates a new Store
func New() *Store {
	return &Store{
		entries:    make(map[string][]parser.Entry),
		containers: []string{},
	}
}

// Add adds an entry to the store
func (s *Store) Add(entry parser.Entry) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if this is a new container
	if _, exists := s.entries[entry.Container]; !exists {
		s.containers = append(s.containers, entry.Container)
		sort.Strings(s.containers)
	}

	s.entries[entry.Container] = append(s.entries[entry.Container], entry)
}

// Containers returns a sorted list of container names
func (s *Store) Containers() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]string, len(s.containers))
	copy(result, s.containers)
	return result
}

// Entries returns all entries for a container
func (s *Store) Entries(container string) []parser.Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries := s.entries[container]
	result := make([]parser.Entry, len(entries))
	copy(result, entries)
	return result
}

// EntryCount returns the number of entries for a container
func (s *Store) EntryCount(container string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.entries[container])
}

// TotalEntries returns the total number of entries across all containers
func (s *Store) TotalEntries() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	total := 0
	for _, entries := range s.entries {
		total += len(entries)
	}
	return total
}
