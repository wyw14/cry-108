package journal

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

func (s *Store) SaveSnapshot(value any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	payload, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	temporary, err := os.CreateTemp(s.dir, "snapshot-*.tmp")
	if err != nil {
		return err
	}
	temporaryName := temporary.Name()
	writeErr := func() error {
		if _, err := temporary.Write(append(payload, '\n')); err != nil {
			return err
		}
		if err := temporary.Sync(); err != nil {
			return err
		}
		if err := temporary.Close(); err != nil {
			return err
		}
		return os.Rename(temporaryName, s.snapPath)
	}()
	if writeErr != nil {
		_ = temporary.Close()
		_ = os.Remove(temporaryName)
	}
	return writeErr
}

func (s *Store) LoadSnapshot(target any) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	payload, err := os.ReadFile(s.snapPath)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if err := json.Unmarshal(payload, target); err != nil {
		return false, err
	}
	return true, nil
}

func (s *Store) SnapshotPath() string {
	return filepath.Clean(s.snapPath)
}
