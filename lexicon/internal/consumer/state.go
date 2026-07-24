package consumer

import (
	"encoding/json"
	"path/filepath"
)

func saveSnapshot(stateRoot, name, snapshotID string) error {
	data, err := json.MarshalIndent(SuccessState{Version: StateVersion, SnapshotID: snapshotID}, "", "  ")
	if err != nil {
		return err
	}
	return writeAtomic(filepath.Join(stateRoot, "consumer-state", name), append(data, '\n'))
}

const StateVersion = 1

type SuccessState struct {
	Version    int    `json:"version"`
	SnapshotID string `json:"snapshot_id"`
}
