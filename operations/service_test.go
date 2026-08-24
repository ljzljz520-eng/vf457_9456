package operations

import (
	"streetlight/domain"
	"streetlight/storage"
	"testing"
)

func TestServiceLifecycle(t *testing.T) {
	db, path, err := storage.OpenEphemeral()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	_ = path
	s := NewService(db)
	if err = s.RegisterCircuit(domain.CircuitLine{ID: "c", Name: "line", Voltage: 220}); err != nil {
		t.Fatal(err)
	}
	if err = s.RegisterPole(domain.StreetlightPole{ID: "p", Code: "P", Location: "road", Active: true}); err != nil {
		t.Fatal(err)
	}
	if err = s.RegisterCrew(domain.Crew{ID: "crew", Name: "crew", Members: []string{"one"}, Active: true}); err != nil {
		t.Fatal(err)
	}
	if _, err = s.ReportFault(domain.FaultReport{PoleID: "p", CircuitID: "c", Reporter: "r", Description: "out", Priority: 2}, "o"); err != nil {
		t.Fatal(err)
	}
	if err = s.AssignCrew(domain.Assignment{WorkOrderID: "o", CrewID: "crew", Dispatcher: "d", Reason: "nearby"}); err != nil {
		t.Fatal(err)
	}
	if err = s.StartRepair("o", "crew", "tech"); err != nil {
		t.Fatal(err)
	}
	if _, err = s.CompleteRepair(domain.RepairUpdate{WorkOrderID: "o", Technician: "tech", Action: "repair", Hours: 1}, "r"); err != nil {
		t.Fatal(err)
	}
	if _, err = s.AcceptOrder(domain.AcceptanceDecision{WorkOrderID: "o", Supervisor: "sup", Passed: true, Findings: "ok"}, "i"); err != nil {
		t.Fatal(err)
	}
	order, err := s.GetWorkOrder("o")
	if err != nil || order.Status != domain.StatusAccepted {
		t.Fatalf("order=%+v err=%v", order, err)
	}
}
