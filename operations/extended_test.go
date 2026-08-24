package operations

import (
	"streetlight/domain"
	"streetlight/storage"
	"testing"
)

func TestAssetHealthAndRepeatReport(t *testing.T) {
	db, path, err := storage.OpenEphemeral()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	_ = path
	s := NewService(db)
	if err = s.RegisterCircuit(domain.CircuitLine{ID: "c", Name: "line", Voltage: 220, PoleIDs: []string{"p"}}); err != nil {
		t.Fatal(err)
	}
	if err = s.RegisterPole(domain.StreetlightPole{ID: "p", Code: "P", Location: "road", Active: true}); err != nil {
		t.Fatal(err)
	}
	if err = s.RegisterControlBox(domain.ControlBox{ID: "b", PoleID: "p", CircuitID: "c", Operational: true}); err != nil {
		t.Fatal(err)
	}
	if _, _, err = s.ReportFaultWithDuplicateCheck(domain.FaultReport{PoleID: "p", CircuitID: "c", Reporter: "desk", Description: "lamp dark", Priority: 3}, "o1"); err != nil {
		t.Fatal(err)
	}
	_, matches, err := s.ReportFaultWithDuplicateCheck(domain.FaultReport{PoleID: "p", CircuitID: "c", Reporter: "desk", Description: " lamp   dark ", Priority: 3}, "o2")
	if err != nil || len(matches) != 1 {
		t.Fatalf("matches=%v err=%v", matches, err)
	}
	if _, err = s.SetCircuitFault("c", "breaker open"); err != nil {
		t.Fatal(err)
	}
	health, err := s.AssetHealthReport()
	if err != nil || len(health) != 2 || health[0].OpenFaults != 2 {
		t.Fatalf("health=%v err=%v", health, err)
	}
}
