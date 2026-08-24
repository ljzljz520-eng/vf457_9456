package dispatch

import (
	"streetlight/domain"
	"streetlight/operations"
	"streetlight/storage"
	"testing"
)

func TestPlannerRanksCrew(t *testing.T) {
	db, path, err := storage.OpenEphemeral()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	_ = path
	s := operations.NewService(db)
	if err = s.RegisterCircuit(domain.CircuitLine{ID: "c", Name: "line", Voltage: 220}); err != nil {
		t.Fatal(err)
	}
	if err = s.RegisterPole(domain.StreetlightPole{ID: "p", Code: "P", Location: "road", Active: true}); err != nil {
		t.Fatal(err)
	}
	for _, crew := range []domain.Crew{{ID: "a", Name: "A", Members: []string{"a"}, Skills: []string{"lamp"}, Active: true}, {ID: "b", Name: "B", Members: []string{"b"}, Skills: []string{"lamp", "box"}, Active: true}} {
		if err = s.RegisterCrew(crew); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = s.ReportFault(domain.FaultReport{PoleID: "p", CircuitID: "c", Reporter: "r", Description: "out", Priority: 3}, "o"); err != nil {
		t.Fatal(err)
	}
	plans, err := NewPlanner(s).Recommend(domain.OrderFilter{Status: domain.StatusReported})
	if err != nil || len(plans) != 1 || plans[0].CrewID != "b" {
		t.Fatalf("plans=%v err=%v", plans, err)
	}
}
