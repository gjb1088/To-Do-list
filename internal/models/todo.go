package models

import (
	"errors"
	"sync"
	"time"
)

// ToDo represents a single task.
type ToDo struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ErrNotFound is returned when a to-do item doesn’t exist.
var ErrNotFound = errors.New("todo not found")

// Store provides thread-safe access to to-do items in memory.
type Store struct {
	sync.RWMutex
	todos  map[int]*ToDo
	nextID int
}

// NewStore initializes a new in-memory store.
func NewStore() *Store {
	return &Store{
		todos:  make(map[int]*ToDo),
		nextID: 1,
	}
}

// Create adds a new to-do and returns it.
func (s *Store) Create(title string) *ToDo {
	s.Lock()
	defer s.Unlock()

	t := &ToDo{
		ID:        s.nextID,
		Title:     title,
		Completed: false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	s.todos[s.nextID] = t
	s.nextID++
	return t
}

// GetAll returns all to-dos in order of creation (simple slice).
func (s *Store) GetAll() []*ToDo {
	s.RLock()
	defer s.RUnlock()

	result := make([]*ToDo, 0, len(s.todos))
	for i := 1; i < s.nextID; i++ {
		if todo, exists := s.todos[i]; exists {
			result = append(result, todo)
		}
	}
	return result
}

// Get retrieves a single to-do by ID.
func (s *Store) Get(id int) (*ToDo, error) {
	s.RLock()
	defer s.RUnlock()

	if todo, ok := s.todos[id]; ok {
		return todo, nil
	}
	return nil, ErrNotFound
}

// Update modifies an existing to-do’s Title and/or Completed fields.
func (s *Store) Update(id int, title string, completed bool) (*ToDo, error) {
	s.Lock()
	defer s.Unlock()

	todo, ok := s.todos[id]
	if !ok {
		return nil, ErrNotFound
	}
	todo.Title = title
	todo.Completed = completed
	todo.UpdatedAt = time.Now()
	return todo, nil
}

// Delete removes a to-do by ID.
func (s *Store) Delete(id int) error {
	s.Lock()
	defer s.Unlock()

	if _, ok := s.todos[id]; !ok {
		return ErrNotFound
	}
	delete(s.todos, id)
	return nil
}

// ClearCompleted removes all tasks where Completed==true.
func (s *Store) ClearCompleted() {
	s.Lock()
	defer s.Unlock()
	for id, todo := range s.todos {
		if todo.Completed {
			delete(s.todos, id)
		}
	}
}
