package operations

import (
	"errors"
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

// maxTransitionRetries bounds the optimistic-concurrency retry loop. bbolt
// commits are serialized, so conflicts are rare; a handful of retries is more
// than enough for any realistic contention, and the bound prevents a runaway
// loop under pathological conditions.
const maxTransitionRetries = 8

// applyTransition is the single atomic entry point for mutating a work order.
// It reads, mutates, and writes the order inside one bbolt transaction via
// UpdateWorkOrder, so the version check and the write cannot be interleaved by
// another operation. If a concurrent writer commits first, UpdateWorkOrder
// returns ErrVersionConflict and applyTransition re-reads and retries against
// the new version — preventing a stale write from clobbering a newer status
// (the dispatch-vs-inspection lost-update bug, where a new acceptance status
// was overwritten by an older dispatch status).
//
// The mutator receives a pointer to the order as currently stored on disk and
// applies the caller's own validation/business rules (each operation preserves
// the rules it always enforced). applyTransition records the change by
// appending a StatusTransition (from the pre-mutation status to the
// post-mutation status) and bumping Version. No transition-graph rule is
// imposed here — callers own their rules, exactly as before.
func (s *Service) applyTransition(orderID string, mutator func(*domain.WorkOrder) (actor, note string, err error)) (domain.WorkOrder, error) {
	for attempt := 0; ; attempt++ {
		current, err := s.DB.GetWorkOrder(orderID)
		if err != nil {
			return domain.WorkOrder{}, err
		}
		if s.hook != nil {
			s.hook("read", orderID)
		}
		expectedVersion := current.Version
		updated, err := s.DB.UpdateWorkOrder(orderID, expectedVersion, func(order domain.WorkOrder) (domain.WorkOrder, error) {
			from := order.Status
			actor, note, e := mutator(&order)
			if e != nil {
				return domain.WorkOrder{}, e
			}
			if actor == "" { // metadata-only refresh, no history entry to record
				return order, nil
			}
			order.History = append(order.History, domain.StatusTransition{Sequence: order.Version + 1, From: from, To: order.Status, Actor: actor, Note: note})
			order.Version++
			return order, nil
		})
		if err == nil {
			if s.hook != nil {
				s.hook("write", orderID)
			}
			return updated, nil
		}
		if !errors.Is(err, storage.ErrVersionConflict) || attempt >= maxTransitionRetries {
			return domain.WorkOrder{}, err
		}
		// Another writer committed first; re-read and retry against its version.
	}
}

func (s *Service) transition(orderID string, to domain.Status, actor, note string) error {
	_, err := s.applyTransition(orderID, func(order *domain.WorkOrder) (string, string, error) {
		if !domain.CanTransition(order.Status, to) {
			return "", "", fmt.Errorf("cannot transition %s from %s to %s", orderID, order.Status, to)
		}
		order.Status = to
		return actor, note, nil
	})
	return err
}

func (s *Service) StartRepair(orderID, crewID, actor string) error {
	crew, err := s.DB.GetCrew(crewID)
	if err != nil {
		return err
	}
	if !crew.Active {
		return fmt.Errorf("crew %s is inactive", crewID)
	}
	// Transition to in-progress and record the crew assignment in one atomic
	// update so a concurrent dispatch/inspection cannot leave the order
	// in-progress with no crew (or vice-versa).
	_, err = s.applyTransition(orderID, func(order *domain.WorkOrder) (string, string, error) {
		if !domain.CanTransition(order.Status, domain.StatusInProgress) {
			return "", "", fmt.Errorf("cannot transition %s from %s to %s", orderID, order.Status, domain.StatusInProgress)
		}
		order.AssignedCrew = crewID
		order.Status = domain.StatusInProgress
		return actor, "repair started", nil
	})
	return err
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
	var target domain.Status
	if decision.Passed {
		target = domain.StatusAccepted
	} else {
		target = domain.StatusRejected
	}
	order, err := s.applyTransition(decision.WorkOrderID, func(order *domain.WorkOrder) (string, string, error) {
		if order.Status != domain.StatusCompleted && order.Status != domain.StatusDispatched {
			return "", "", fmt.Errorf("order %s is not ready for inspection", order.ID)
		}
		order.Status = target
		return decision.Supervisor, decision.Findings, nil
	})
	if err != nil {
		return domain.Inspection{}, err
	}
	inspection := domain.Inspection{ID: inspectionID, WorkOrderID: order.ID, Supervisor: decision.Supervisor, Passed: decision.Passed, Findings: decision.Findings, Sequence: order.Version}
	return inspection, s.DB.SaveInspection(inspection)
}

func (s *Service) ReopenRejected(orderID, actor, note string) error {
	return s.transition(orderID, domain.StatusInProgress, actor, note)
}
