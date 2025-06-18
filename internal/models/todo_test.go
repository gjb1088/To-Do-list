package models

import "testing"

func TestStoreCreateAndGet(t *testing.T) {
	s := NewStore()
	todo := s.Create("write tests")
	if todo.ID != 1 {
		t.Fatalf("expected ID 1 got %d", todo.ID)
	}
	if todo.Title != "write tests" || todo.Completed {
		t.Fatalf("unexpected todo %+v", todo)
	}
	fetched, err := s.Get(todo.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetched != todo {
		t.Fatalf("fetched todo does not match")
	}
	all := s.GetAll()
	if len(all) != 1 || all[0] != todo {
		t.Fatalf("GetAll returned %#v", all)
	}
}

func TestStoreUpdate(t *testing.T) {
	s := NewStore()
	todo := s.Create("initial")
	updated, err := s.Update(todo.ID, "changed", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Title != "changed" || !updated.Completed {
		t.Fatalf("update failed: %+v", updated)
	}
	fetched, _ := s.Get(todo.ID)
	if fetched.Title != "changed" || !fetched.Completed {
		t.Fatalf("Get after update mismatch: %+v", fetched)
	}
}

func TestStoreDelete(t *testing.T) {
	s := NewStore()
	a := s.Create("a")
	b := s.Create("b")
	if err := s.Delete(a.ID); err != nil {
		t.Fatalf("unexpected delete error: %v", err)
	}
	if _, err := s.Get(a.ID); err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	all := s.GetAll()
	if len(all) != 1 || all[0] != b {
		t.Fatalf("unexpected todos after delete: %#v", all)
	}
}

func TestStoreClearCompleted(t *testing.T) {
	s := NewStore()
	a := s.Create("a")
	b := s.Create("b")
	c := s.Create("c")
	s.Update(a.ID, a.Title, true)
	s.Update(c.ID, c.Title, true)
	s.ClearCompleted()
	all := s.GetAll()
	if len(all) != 1 || all[0] != b {
		t.Fatalf("expected only b remaining, got %#v", all)
	}
}
