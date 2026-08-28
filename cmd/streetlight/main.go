package main

import (
	"fmt"
	"os"
	"streetlight/domain"
	"streetlight/inspection"
	"streetlight/operations"
	"streetlight/reporting"
	"streetlight/storage"
	"streetlight/workflow"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("streetlight commands: demo, report")
		return
	}
	db, path, err := storage.OpenEphemeral()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer func() { _ = db.Close(); _ = os.Remove(path) }()
	service := operations.NewService(db)
	if err = workflow.Seed(service); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	switch os.Args[1] {
	case "demo":
		err = runDemo(service)
	case "report":
		err = runReport(service)
	default:
		fmt.Fprintln(os.Stderr, "unknown command")
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runDemo(service *operations.Service) error {
	coordinator := workflow.NewCoordinator(service)
	plan, err := coordinator.RecordAndPlan(domain.FaultReport{PoleID: "pole-1", CircuitID: "circuit-1", Reporter: "control-room", Description: "lamp is dark", Priority: 4}, "order-demo")
	if err != nil {
		return err
	}
	if _, err = coordinator.DispatchAndRepair(plan, "dispatcher", domain.RepairUpdate{WorkOrderID: "order-demo", Technician: "Ava", Action: "replace driver", Material: "LED driver", Hours: 2}, "repair-demo"); err != nil {
		return err
	}
	result, err := coordinator.InspectAndReport("order-demo", "inspection-demo", "supervisor", inspection.Checklist{PowerRestored: true, LampAligned: true, CabinetLocked: true, AreaSafe: true}, "passed")
	if err != nil {
		return err
	}
	fmt.Print(result)
	return nil
}

func runReport(service *operations.Service) error {
	report, err := reporting.Build(service, domain.OrderFilter{})
	if err != nil {
		return err
	}
	fmt.Print(reporting.Render(report))
	return nil
}
