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
	order, err := s.DB.GetWorkOrder(assignment.WorkOrderID)
	if err != nil {
		return err
	}
	if s.hook != nil {
		s.hook("read", assignment.WorkOrderID)
	}
	if order.Status == domain.StatusAccepted {
		return fmt.Errorf("accepted order cannot be dispatched")
	}
	order.AssignedCrew = assignment.CrewID
	order.Status = domain.StatusDispatched
	order.Version++
	order.History = append(order.History, domain.StatusTransition{Sequence: order.Version, From: order.History[len(order.History)-1].To, To: domain.StatusDispatched, Actor: assignment.Dispatcher, Note: assignment.Reason})
	if s.hook != nil {
		s.hook("write", assignment.WorkOrderID)
	}
	return s.DB.SaveWorkOrder(order)
}

func (s *Service) CancelOrder(orderID, actor, reason string) error {
	order, err := s.DB.GetWorkOrder(orderID)
	if err != nil {
		return err
	}
	if order.Status == domain.StatusAccepted {
		return fmt.Errorf("accepted order cannot be cancelled")
	}
	if reason == "" {
		return fmt.Errorf("cancellation reason is required")
	}
	order.Description = order.Description + " [cancelled: " + reason + "]"
	order.Version++
	order.History = append(order.History, domain.StatusTransition{Sequence: order.Version, From: order.Status, To: order.Status, Actor: actor, Note: "cancelled"})
	return s.DB.SaveWorkOrder(order)
}

func (s *Service) EscalatePriority(orderID, actor string) (domain.WorkOrder, error) {
	order, err := s.DB.GetWorkOrder(orderID)
	if err != nil {
		return order, err
	}
	if order.Priority >= 5 {
		return order, fmt.Errorf("order already has highest priority")
	}
	order.Priority++
	order.Version++
	order.History = append(order.History, domain.StatusTransition{Sequence: order.Version, From: order.Status, To: order.Status, Actor: actor, Note: "priority escalated"})
	return order, s.DB.SaveWorkOrder(order)
}
