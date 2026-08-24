package workflow

import (
	"os"
	"streetlight/domain"
	"streetlight/inspection"
	"streetlight/operations"
	"streetlight/storage"
	"testing"
)

func testService(t *testing.T) (*operations.Service, func()) {
	t.Helper()
	db, path, err := storage.OpenEphemeral()
	if err != nil {
		t.Fatal(err)
	}
	service := operations.NewService(db)
	if err = Seed(service); err != nil {
		t.Fatal(err)
	}
	return service, func() { _ = db.Close(); _ = removeFile(path) }
}

func removeFile(path string) error { return os.Remove(path) }

func TestWorkflowOne(t *testing.T) {
	service, done := testService(t)
	defer done()
	c := NewCoordinator(service)
	plan, err := c.RecordAndPlan(domain.FaultReport{PoleID: "pole-1", CircuitID: "circuit-1", Reporter: "desk", Description: "dark lamp", Priority: 4}, "order-1")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = c.DispatchAndRepair(plan, "dispatcher", domain.RepairUpdate{WorkOrderID: "order-1", Technician: "Ava", Action: "replace lamp", Material: "lamp", Hours: 2}, "repair-1"); err != nil {
		t.Fatal(err)
	}
	result, err := c.InspectAndReport("order-1", "inspection-1", "supervisor", inspection.Checklist{PowerRestored: true, LampAligned: true, CabinetLocked: true, AreaSafe: true}, "passed")
	if err != nil || result == "" {
		t.Fatalf("workflow failed: %v", err)
	}
	order, err := service.GetWorkOrder("order-1")
	if err != nil || order.Status != domain.StatusAccepted {
		t.Fatalf("unexpected order: %+v %v", order, err)
	}
}

func TestWorkflowTwo(t *testing.T) {
	service, done := testService(t)
	defer done()
	if _, err := service.ReportFault(domain.FaultReport{PoleID: "pole-2", CircuitID: "circuit-1", Reporter: "desk", Description: "flicker", Priority: 2}, "order-2"); err != nil {
		t.Fatal(err)
	}
	planner := NewCoordinator(service).Planner
	plans, err := planner.Recommend(domain.OrderFilter{Status: domain.StatusReported})
	if err != nil || len(plans) != 1 {
		t.Fatalf("plans=%v err=%v", plans, err)
	}
	if err = planner.Dispatch(plans[0], "dispatcher"); err != nil {
		t.Fatal(err)
	}
	orders, err := service.ListOrders(domain.OrderFilter{CrewID: "crew-1"})
	if err != nil || len(orders) != 1 {
		t.Fatalf("orders=%v err=%v", orders, err)
	}
}

func TestWorkflowThree(t *testing.T) {
	service, done := testService(t)
	defer done()
	c := NewCoordinator(service)
	plan, err := c.RecordAndPlan(domain.FaultReport{PoleID: "pole-1", CircuitID: "circuit-1", Reporter: "desk", Description: "cabinet alarm", Priority: 5}, "order-3")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = c.DispatchAndRepair(plan, "dispatcher", domain.RepairUpdate{WorkOrderID: "order-3", Technician: "Bo", Action: "tighten terminal", Material: "terminal", Hours: 1}, "repair-3"); err != nil {
		t.Fatal(err)
	}
	if _, err = c.Inspector.Review("order-3", "inspection-3", "supervisor", inspection.Checklist{PowerRestored: false, LampAligned: true, CabinetLocked: true, AreaSafe: true}, "power still unstable"); err != nil {
		t.Fatal(err)
	}
	order, err := service.GetWorkOrder("order-3")
	if err != nil || order.Status != domain.StatusRejected {
		t.Fatalf("unexpected status %s %v", order.Status, err)
	}
}
