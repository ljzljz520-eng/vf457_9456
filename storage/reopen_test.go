package storage

import (
	"path/filepath"
	"streetlight/domain"
	"testing"
)

func TestPersistenceSurvivesReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "records.db")
	db, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if err = db.SaveStreetlightPole(domain.StreetlightPole{ID: "p", Code: "P-1", Location: "yard", Active: true}); err != nil {
		t.Fatal(err)
	}
	if err = db.SaveCircuitLine(domain.CircuitLine{ID: "c", Name: "feeder", Voltage: 220}); err != nil {
		t.Fatal(err)
	}
	if err = db.SaveControlBox(domain.ControlBox{ID: "b", PoleID: "p", CircuitID: "c", Address: "yard", Operational: true}); err != nil {
		t.Fatal(err)
	}
	if err = db.SaveCrew(domain.Crew{ID: "crew", Name: "crew", Members: []string{"one"}, Active: true}); err != nil {
		t.Fatal(err)
	}
	if err = db.SaveWorkOrder(domain.WorkOrder{ID: "o", PoleID: "p", CircuitID: "c", Description: "test", Priority: 1, Status: domain.StatusReported, History: []domain.StatusTransition{{To: domain.StatusReported}}}); err != nil {
		t.Fatal(err)
	}
	if err = db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err = db.GetStreetlightPole("p"); err != nil {
		t.Fatal(err)
	}
	if _, err = db.GetControlBox("b"); err != nil {
		t.Fatal(err)
	}
	if _, err = db.GetCircuitLine("c"); err != nil {
		t.Fatal(err)
	}
	if _, err = db.GetCrew("crew"); err != nil {
		t.Fatal(err)
	}
	if _, err = db.GetWorkOrder("o"); err != nil {
		t.Fatal(err)
	}
}

func TestReferencesRejectUnknownCrew(t *testing.T) {
	db, path, err := OpenEphemeral()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	_ = path
	if err = db.SaveWorkOrder(domain.WorkOrder{ID: "o", PoleID: "p", Priority: 1, Status: domain.StatusReported, History: []domain.StatusTransition{{To: domain.StatusReported}}, AssignedCrew: "missing"}); err != nil {
		t.Fatal(err)
	}
	if err = db.ValidateReferences(); err == nil {
		t.Fatal("expected reference error")
	}
}
