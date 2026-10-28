package main

import (
	"context"
	"reflect"
	"testing"

	secretmanager "github.com/sacloud/sacloud-sdk-go/api/secretmanager"
)

type recordingDeleter struct {
	secretmanager.SecretAPI
	deleted []string
}

func (r *recordingDeleter) Delete(_ context.Context, name string) error {
	r.deleted = append(r.deleted, name)
	return nil
}

func TestDeleteOrphansDeletesOnlyUnsyncedSecrets(t *testing.T) {
	existing := map[string]bool{"a": true, "b/x": true, "c": true, "d": true}
	synced := map[string]bool{"a": true, "c": true, "new": true}

	op := &recordingDeleter{}
	if err := deleteOrphans(op, existing, synced, false); err != nil {
		t.Fatal(err)
	}

	want := []string{"b/x", "d"}
	if !reflect.DeepEqual(op.deleted, want) {
		t.Errorf("deleted %q, want %q", op.deleted, want)
	}
}

func TestDeleteOrphansDryRunDeletesNothing(t *testing.T) {
	existing := map[string]bool{"a": true, "b": true}
	synced := map[string]bool{"a": true}

	op := &recordingDeleter{}
	if err := deleteOrphans(op, existing, synced, true); err != nil {
		t.Fatal(err)
	}

	if len(op.deleted) != 0 {
		t.Errorf("deleted %q in dry-run, want nothing", op.deleted)
	}
}
