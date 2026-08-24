package journal

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"

	"github.com/google/uuid"
	"github.com/wyw14/fabchem/internal/model"
)

type Store struct {
	mu       sync.Mutex
	dir      string
	logPath  string
	snapPath string
}

func Open(dir string) (*Store, error) {
	if dir == "" {
		return nil, errors.New("journal directory is required")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return &Store{
		dir:      dir,
		logPath:  filepath.Join(dir, "events.jsonl"),
		snapPath: filepath.Join(dir, "snapshot.json"),
	}, nil
}

func (s *Store) Append(ctx context.Context, event model.Event) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if event.ID == "" {
		event.ID = uuid.NewString()
	}
	file, err := os.OpenFile(s.logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(file)
	encodeErr := encoder.Encode(event)
	syncErr := file.Sync()
	closeErr := file.Close()
	return errors.Join(encodeErr, syncErr, closeErr)
}

func (s *Store) ReadAll(ctx context.Context) ([]model.Event, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	file, err := os.Open(s.logPath)
	if errors.Is(err, os.ErrNotExist) {
		return []model.Event{}, nil
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()
	events := make([]model.Event, 0)
	scanner := bufio.NewScanner(file)
	buffer := make([]byte, 0, 64*1024)
	scanner.Buffer(buffer, 1024*1024)
	for scanner.Scan() {
		var event model.Event
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, scanner.Err()
}
