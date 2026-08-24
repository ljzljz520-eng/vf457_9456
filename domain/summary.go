package domain

type Dashboard struct {
	TotalPoles      int
	ActivePoles     int
	OpenOrders      int
	CompletedOrders int
	AcceptedOrders  int
	RejectedOrders  int
	FaultedCircuits int
	AvailableCrews  int
}

type OrderFilter struct {
	Status          Status
	CrewID          string
	CircuitID       string
	MinimumPriority int
}

func (f OrderFilter) Matches(w WorkOrder) bool {
	if f.Status != "" && w.Status != f.Status {
		return false
	}
	if f.CrewID != "" && w.AssignedCrew != f.CrewID {
		return false
	}
	if f.CircuitID != "" && w.CircuitID != f.CircuitID {
		return false
	}
	if f.MinimumPriority > 0 && w.Priority < f.MinimumPriority {
		return false
	}
	return true
}
