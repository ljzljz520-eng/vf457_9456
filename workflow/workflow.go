package workflow

import (
	"fmt"
	"streetlight/dispatch"
	"streetlight/domain"
	"streetlight/inspection"
	"streetlight/operations"
	"streetlight/reporting"
)

type Coordinator struct {
	Service   *operations.Service
	Planner   *dispatch.Planner
	Inspector *inspection.Inspector
}

func NewCoordinator(service *operations.Service) *Coordinator {
	return &Coordinator{Service: service, Planner: dispatch.NewPlanner(service), Inspector: inspection.NewInspector(service)}
}

func (c *Coordinator) RecordAndPlan(report domain.FaultReport, orderID string) (dispatch.PlanItem, error) {
	if _, err := c.Service.ReportFault(report, orderID); err != nil {
		return dispatch.PlanItem{}, err
	}
	plans, err := c.Planner.Recommend(dispatch.BuildFilter(domain.StatusReported, report.CircuitID, 0))
	if err != nil {
		return dispatch.PlanItem{}, err
	}
	if len(plans) == 0 {
		return dispatch.PlanItem{}, fmt.Errorf("no active crew available")
	}
	return plans[0], nil
}

func (c *Coordinator) DispatchAndRepair(plan dispatch.PlanItem, dispatcher string, update domain.RepairUpdate, recordID string) (domain.RepairRecord, error) {
	if err := c.Planner.Dispatch(plan, dispatcher); err != nil {
		return domain.RepairRecord{}, err
	}
	if err := c.Service.StartRepair(update.WorkOrderID, plan.CrewID, update.Technician); err != nil {
		return domain.RepairRecord{}, err
	}
	return c.Service.CompleteRepair(update, recordID)
}

func (c *Coordinator) InspectAndReport(orderID, inspectionID, supervisor string, checklist inspection.Checklist, findings string) (string, error) {
	item, err := c.Inspector.Review(orderID, inspectionID, supervisor, checklist, findings)
	if err != nil {
		return "", err
	}
	report, err := reporting.Build(c.Service, domain.OrderFilter{})
	if err != nil {
		return "", err
	}
	return reporting.StatusLabel(itemStatus(item)) + "\n" + reporting.Render(report), nil
}

func itemStatus(item domain.Inspection) domain.Status {
	if item.Passed {
		return domain.StatusAccepted
	}
	return domain.StatusRejected
}
