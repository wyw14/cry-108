package journal_test

import (
	"context"
	"testing"

	"github.com/wyw14/fabchem/internal/journal"
	"github.com/wyw14/fabchem/internal/model"
)

func TestJournalAppendReplayAndSnapshot(t *testing.T) {
	store, err := journal.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := store.Append(ctx, model.NewEvent("facility.ready", "fabchem", map[string]any{"ready": true})); err != nil {
		t.Fatal(err)
	}
	count := 0
	replayed, err := store.Replay(ctx, func(event model.Event) error {
		count++
		return nil
	})
	if err != nil || replayed != 1 || count != 1 {
		t.Fatalf("unexpected replay: count=%d replayed=%d err=%v", count, replayed, err)
	}
	if err := store.SaveSnapshot(map[string]any{"ready": true}); err != nil {
		t.Fatal(err)
	}
	var snapshot map[string]any
	ok, err := store.LoadSnapshot(&snapshot)
	if err != nil || !ok || snapshot["ready"] != true {
		t.Fatalf("unexpected snapshot: ok=%v value=%v err=%v", ok, snapshot, err)
	}
	if store.Directory() == "" || store.SnapshotPath() == "" {
		t.Fatal("journal paths must be available to runtime operations")
	}
}
