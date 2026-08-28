package storage

import (
	"streetlight/domain"
	"testing"
)

func TestSnapshotRoundTrip(t *testing.T) {
	db, path, err := OpenEphemeral()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	_ = path
	if err = db.SaveStreetlightPole(domain.StreetlightPole{ID: "p", Code: "P", Location: "road", Active: true}); err != nil {
		t.Fatal(err)
	}
	if err = db.SaveCircuitLine(domain.CircuitLine{ID: "c", Name: "line", Voltage: 220}); err != nil {
		t.Fatal(err)
	}
	data, err := db.SnapshotJSON()
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := DecodeSnapshot(data)
	if err != nil || len(snapshot.Poles) != 1 {
		t.Fatalf("snapshot=%+v err=%v", snapshot, err)
	}
	if err = db.RestoreSnapshot(snapshot); err != nil {
		t.Fatal(err)
	}
	if _, err = db.GetStreetlightPole("p"); err != nil {
		t.Fatal(err)
	}
}
