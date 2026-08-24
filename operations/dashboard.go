package operations

import "streetlight/domain"

func (s *Service) BuildDashboard() (domain.Dashboard, error) {
	inv, err := s.DB.Inventory()
	if err != nil {
		return domain.Dashboard{}, err
	}
	result := domain.Dashboard{TotalPoles: len(inv.Poles), AvailableCrews: 0}
	for _, pole := range inv.Poles {
		if pole.Active {
			result.ActivePoles++
		}
	}
	for _, crew := range inv.Crews {
		if crew.Active {
			result.AvailableCrews++
		}
	}
	for _, circuit := range inv.Circuits {
		if circuit.Faulted {
			result.FaultedCircuits++
		}
	}
	for _, order := range inv.Orders {
		switch order.Status {
		case domain.StatusAccepted:
			result.AcceptedOrders++
		case domain.StatusCompleted:
			result.CompletedOrders++
		case domain.StatusRejected:
			result.RejectedOrders++
		default:
			result.OpenOrders++
		}
	}
	return result, nil
}

func (s *Service) RepairHistory(orderID string) ([]domain.StatusTransition, error) {
	order, err := s.DB.GetWorkOrder(orderID)
	if err != nil {
		return nil, err
	}
	history := make([]domain.StatusTransition, len(order.History))
	copy(history, order.History)
	return history, nil
}
