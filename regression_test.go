package streetlight

import (
	"streetlight/domain"
	"streetlight/operations"
	"streetlight/storage"
	"sync"
	"testing"
)

func TestStreetlightStatusKeepsLatestTransition(t *testing.T) {
	db, path, err := storage.OpenEphemeral()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	_ = path
	base := operations.NewService(db)
	if err = base.RegisterCircuit(domain.CircuitLine{ID: "c", Name: "line", Voltage: 220}); err != nil {
		t.Fatal(err)
	}
	if err = base.RegisterPole(domain.StreetlightPole{ID: "p", Code: "P", Location: "road", Active: true}); err != nil {
		t.Fatal(err)
	}
	if err = base.RegisterCrew(domain.Crew{ID: "crew", Name: "crew", Members: []string{"one"}, Active: true}); err != nil {
		t.Fatal(err)
	}
	if _, err = base.ReportFault(domain.FaultReport{PoleID: "p", CircuitID: "c", Reporter: "desk", Description: "out", Priority: 4}, "o"); err != nil {
		t.Fatal(err)
	}
	if err = base.AssignCrew(domain.Assignment{WorkOrderID: "o", CrewID: "crew", Dispatcher: "d", Reason: "first"}); err != nil {
		t.Fatal(err)
	}
	if err = base.StartRepair("o", "crew", "tech"); err != nil {
		t.Fatal(err)
	}
	if _, err = base.CompleteRepair(domain.RepairUpdate{WorkOrderID: "o", Technician: "tech", Action: "repair", Hours: 1}, "r"); err != nil {
		t.Fatal(err)
	}
	reads := make(chan struct{})
	var readCount int
	var mu sync.Mutex
	acceptDone := make(chan struct{})
	markRead := func(stage, id string) {
		if stage != "read" {
			return
		}
		mu.Lock()
		readCount++
		n := readCount
		mu.Unlock()
		if n == 2 {
			close(reads)
		}
		<-reads
	}
	dispatchService := operations.NewService(db).WithTransitionHook(func(stage, id string) {
		if stage == "read" {
			markRead(stage, id)
		} else if stage == "write" {
			<-acceptDone
		}
	})
	acceptService := operations.NewService(db).WithTransitionHook(markRead)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		_ = dispatchService.AssignCrew(domain.Assignment{WorkOrderID: "o", CrewID: "crew", Dispatcher: "d", Reason: "late route"})
	}()
	go func() {
		defer wg.Done()
		_, _ = acceptService.AcceptOrder(domain.AcceptanceDecision{WorkOrderID: "o", Supervisor: "sup", Passed: true, Findings: "passed"}, "i")
		close(acceptDone)
	}()
	wg.Wait()
	order, err := base.GetWorkOrder("o")
	if err != nil {
		t.Fatal(err)
	}
	if order.Status != domain.StatusAccepted {
		t.Fatalf("latest status lost: %s", order.Status)
	}
}
