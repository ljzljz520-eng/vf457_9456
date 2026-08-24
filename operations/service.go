package operations

import (
	"fmt"
	"streetlight/domain"
	"streetlight/storage"
)

type TransitionHook func(stage, orderID string)

type Service struct {
	DB   *storage.DB
	hook TransitionHook
}

func NewService(db *storage.DB) *Service { return &Service{DB: db} }

func (s *Service) WithTransitionHook(h TransitionHook) *Service { s.hook = h; return s }

func (s *Service) RegisterPole(pole domain.StreetlightPole) error {
	if err := pole.Validate(); err != nil {
		return err
	}
	if _, err := s.DB.GetStreetlightPole(pole.ID); err == nil {
		return fmt.Errorf("pole %s already exists", pole.ID)
	}
	return s.DB.SaveStreetlightPole(pole)
}

func (s *Service) RegisterControlBox(box domain.ControlBox) error {
	if err := box.Validate(); err != nil {
		return err
	}
	if _, err := s.DB.GetStreetlightPole(box.PoleID); err != nil {
		return fmt.Errorf("pole reference: %w", err)
	}
	if _, err := s.DB.GetCircuitLine(box.CircuitID); err != nil {
		return fmt.Errorf("circuit reference: %w", err)
	}
	return s.DB.SaveControlBox(box)
}

func (s *Service) RegisterCircuit(circuit domain.CircuitLine) error {
	if err := circuit.Validate(); err != nil {
		return err
	}
	return s.DB.SaveCircuitLine(circuit)
}

func (s *Service) RegisterCrew(crew domain.Crew) error {
	if err := crew.Validate(); err != nil {
		return err
	}
	return s.DB.SaveCrew(crew)
}

func (s *Service) ReportFault(report domain.FaultReport, orderID string) (domain.WorkOrder, error) {
	report = report.Normalize()
	if report.PoleID == "" || report.Description == "" || report.Reporter == "" {
		return domain.WorkOrder{}, fmt.Errorf("fault report is incomplete")
	}
	if _, err := s.DB.GetStreetlightPole(report.PoleID); err != nil {
		return domain.WorkOrder{}, fmt.Errorf("pole reference: %w", err)
	}
	if report.CircuitID != "" {
		if _, err := s.DB.GetCircuitLine(report.CircuitID); err != nil {
			return domain.WorkOrder{}, fmt.Errorf("circuit reference: %w", err)
		}
	}
	order := domain.WorkOrder{ID: orderID, PoleID: report.PoleID, CircuitID: report.CircuitID, Description: report.Description, Priority: report.Priority, Status: domain.StatusReported, Version: 1}
	order.History = append(order.History, domain.StatusTransition{Sequence: 1, From: "", To: domain.StatusReported, Actor: report.Reporter, Note: "fault recorded"})
	if err := order.Validate(); err != nil {
		return domain.WorkOrder{}, err
	}
	return order, s.DB.SaveWorkOrder(order)
}

func (s *Service) GetWorkOrder(id string) (domain.WorkOrder, error) { return s.DB.GetWorkOrder(id) }

func (s *Service) ListOrders(filter domain.OrderFilter) ([]domain.WorkOrder, error) {
	items, err := s.DB.ListWorkOrders()
	if err != nil {
		return nil, err
	}
	out := make([]domain.WorkOrder, 0, len(items))
	for _, item := range items {
		if filter.Matches(item) {
			out = append(out, item)
		}
	}
	storage.SortWorkOrdersByPriority(out)
	return out, nil
}

func (s *Service) transition(orderID string, to domain.Status, actor, note string) error {
	order, err := s.DB.GetWorkOrder(orderID)
	if err != nil {
		return err
	}
	if s.hook != nil {
		s.hook("read", orderID)
	}
	if !domain.CanTransition(order.Status, to) {
		return fmt.Errorf("cannot transition %s from %s to %s", orderID, order.Status, to)
	}
	order.Status = to
	order.Version++
	order.History = append(order.History, domain.StatusTransition{Sequence: order.Version, From: order.History[len(order.History)-1].To, To: to, Actor: actor, Note: note})
	if s.hook != nil {
		s.hook("write", orderID)
	}
	return s.DB.SaveWorkOrder(order)
}

func (s *Service) StartRepair(orderID, crewID, actor string) error {
	crew, err := s.DB.GetCrew(crewID)
	if err != nil {
		return err
	}
	if !crew.Active {
		return fmt.Errorf("crew %s is inactive", crewID)
	}
	if err = s.transition(orderID, domain.StatusInProgress, actor, "repair started"); err != nil {
		return err
	}
	order, err := s.DB.GetWorkOrder(orderID)
	if err != nil {
		return err
	}
	order.AssignedCrew = crewID
	return s.DB.SaveWorkOrder(order)
}

func (s *Service) CompleteRepair(update domain.RepairUpdate, recordID string) (domain.RepairRecord, error) {
	if !update.IsComplete() {
		return domain.RepairRecord{}, fmt.Errorf("repair update is incomplete")
	}
	if err := s.transition(update.WorkOrderID, domain.StatusCompleted, update.Technician, "repair completed"); err != nil {
		return domain.RepairRecord{}, err
	}
	record := domain.RepairRecord{ID: recordID, WorkOrderID: update.WorkOrderID, Action: update.Action, Material: update.Material, Hours: update.Hours, CompletedBy: update.Technician}
	return record, s.DB.SaveRepairRecord(record)
}

func (s *Service) AcceptOrder(decision domain.AcceptanceDecision, inspectionID string) (domain.Inspection, error) {
	if !decision.IsComplete() {
		return domain.Inspection{}, fmt.Errorf("acceptance decision is incomplete")
	}
	order, err := s.DB.GetWorkOrder(decision.WorkOrderID)
	if err != nil {
		return domain.Inspection{}, err
	}
	if s.hook != nil {
		s.hook("read", decision.WorkOrderID)
	}
	if order.Status != domain.StatusCompleted && order.Status != domain.StatusDispatched {
		return domain.Inspection{}, fmt.Errorf("order %s is not ready for inspection", order.ID)
	}
	if decision.Passed {
		order.Status = domain.StatusAccepted
	} else {
		order.Status = domain.StatusRejected
	}
	order.Version++
	order.History = append(order.History, domain.StatusTransition{Sequence: order.Version, From: order.History[len(order.History)-1].To, To: order.Status, Actor: decision.Supervisor, Note: decision.Findings})
	if s.hook != nil {
		s.hook("write", decision.WorkOrderID)
	}
	if err = s.DB.SaveWorkOrder(order); err != nil {
		return domain.Inspection{}, err
	}
	inspection := domain.Inspection{ID: inspectionID, WorkOrderID: order.ID, Supervisor: decision.Supervisor, Passed: decision.Passed, Findings: decision.Findings, Sequence: order.Version}
	return inspection, s.DB.SaveInspection(inspection)
}

func (s *Service) ReopenRejected(orderID, actor, note string) error {
	return s.transition(orderID, domain.StatusInProgress, actor, note)
}
