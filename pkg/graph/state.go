// Package graph provides state management for graph execution
package graph

import (
	"encoding/json"
	"sync"
)

// State holds accumulated data as the graph executes
type State struct {
	mu   sync.RWMutex
	data map[string]interface{}
}

// NewState creates a new empty state
func NewState() *State {
	return &State{
		data: make(map[string]interface{}),
	}
}

// Set stores a value in the state
func (s *State) Set(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}

// Get retrieves a value from the state
func (s *State) Get(key string) (interface{}, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.data[key]
	return val, ok
}

// GetString retrieves a string value
func (s *State) GetString(key string) (string, bool) {
	val, ok := s.Get(key)
	if !ok {
		return "", false
	}
	str, ok := val.(string)
	return str, ok
}

// GetStringSlice retrieves a string slice
func (s *State) GetStringSlice(key string) ([]string, bool) {
	val, ok := s.Get(key)
	if !ok {
		return nil, false
	}
	slice, ok := val.([]string)
	return slice, ok
}

// GetJSON retrieves a json.RawMessage value
func (s *State) GetJSON(key string) (json.RawMessage, bool) {
	val, ok := s.Get(key)
	if !ok {
		return nil, false
	}
	data, ok := val.(json.RawMessage)
	return data, ok
}

// Has checks if a key exists
func (s *State) Has(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.data[key]
	return ok
}

// Keys returns all state keys
func (s *State) Keys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	keys := make([]string, 0, len(s.data))
	for k := range s.data {
		keys = append(keys, k)
	}
	return keys
}

// Clone creates a shallow copy of the state
func (s *State) Clone() *State {
	s.mu.RLock()
	defer s.mu.RUnlock()

	clone := NewState()
	for k, v := range s.data {
		clone.data[k] = v
	}
	return clone
}
