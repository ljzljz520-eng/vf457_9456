package operations

import (
	"fmt"
	"streetlight/domain"
)

func (s *Service) AssignCrew(assignment domain.Assignment) error {
	if !assignment.IsComplete() {
		return fmt.Errorf("assignment is incomplete")
	}
	crew, err := s.DB.GetCrew(assignment.CrewID)
	if err != nil {
		return err
	}
	if !crew.Active {
		return fmt.Errorf("crew %s is inactive", crew.ID)
	}
	// Dispatch and crew assignment run as one atomic, version-checked update so
	// a concurrent inspection cannot be silently overwritten by this dispatch
	// (and vice-versa). If AcceptOrder committed first, the version check
	// fails and we retry against the now-accepted order — which then rejects
	// dispatch instead of clobbering the acceptance status.
	_, err = s.applyTransition(assignment.WorkOrderID, func(order *domain.WorkOrder) (string, string, error) {
		if order.Status == domain.StatusAccepted {
			return "", "", fmt.Errorf("accepted order cannot be dispatched")
		}
		if order.Status == domain.StatusDispatched {
			order.AssignedCrew = assignment.CrewID
			return "", "", nil // already dispatched; only refresh crew
		}
		order.AssignedCrew = assignment.CrewID
		order.Status = domain.StatusDispatched
		return assignment.Dispatcher, assignment.Reason, nil
	})
	return err
}

func (s *Service) CancelOrder(orderID, actor, reason string) error {
	if reason == "" {
		return fmt.Errorf("cancellation reason is required")
	}
	_, err := s.applyTransition(orderID, func(order *domain.WorkOrder) (string, string, error) {
		if order.Status == domain.StatusAccepted {
			return "", "", fmt.Errorf("accepted order cannot be cancelled")
		}
		order.Description = order.Description + " [cancelled: " + reason + "]"
		return actor, "cancelled", nil // status unchanged; recorded with from == to
	})
	return err
}

func (s *Service) EscalatePriority(orderID, actor string) (domain.WorkOrder, error) {
	order, err := s.applyTransition(orderID, func(order *domain.WorkOrder) (string, string, error) {
		if order.Priority >= 5 {
			return "", "", fmt.Errorf("order already has highest priority")
		}
		order.Priority++
		return actor, "priority escalated", nil
	})
	return order, err
}
