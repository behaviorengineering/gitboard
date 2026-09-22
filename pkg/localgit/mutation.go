package localgit

import "sync"

// MutationLocks serializes mutating git operations per common git directory
// so unrelated repositories stay parallel.
type MutationLocks struct {
	mu    sync.Mutex
	locks map[string]*sync.Mutex
}

// NewMutationLocks returns an empty lock table.
func NewMutationLocks() *MutationLocks {
	return &MutationLocks{locks: map[string]*sync.Mutex{}}
}

// Lock acquires the mutex for commonDir and returns an unlock function.
func (m *MutationLocks) Lock(commonDir string) (unlock func()) {
	if m == nil {
		return func() {}
	}
	if commonDir == "" {
		commonDir = "_"
	}
	m.mu.Lock()
	if m.locks == nil {
		m.locks = map[string]*sync.Mutex{}
	}
	lk, ok := m.locks[commonDir]
	if !ok {
		lk = &sync.Mutex{}
		m.locks[commonDir] = lk
	}
	m.mu.Unlock()
	lk.Lock()
	return lk.Unlock
}
