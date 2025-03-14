package randname

import "sync"

type Store struct {
	lock  sync.Mutex
	store map[string]struct{}
}

func New(names []string) *Store {
	store := make(map[string]struct{})
	for _, n := range names {
		store[n] = struct{}{}
	}
	return &Store{store: store}
}

func (s *Store) Put(name string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.store[name] = struct{}{}
}

func (s *Store) Pop() string {
	s.lock.Lock()
	defer s.lock.Unlock()
	for name := range s.store {
		return name
	}
	return ""
}
