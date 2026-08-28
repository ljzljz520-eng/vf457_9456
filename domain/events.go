package domain

type FaultReport struct {
	PoleID      string
	CircuitID   string
	Reporter    string
	Description string
	Priority    int
}

type Assignment struct {
	WorkOrderID string
	CrewID      string
	Dispatcher  string
	Reason      string
}

type RepairUpdate struct {
	WorkOrderID string
	Technician  string
	Action      string
	Material    string
	Hours       int
}

type AcceptanceDecision struct {
	WorkOrderID string
	Supervisor  string
	Passed      bool
	Findings    string
}

func (r FaultReport) Normalize() FaultReport {
	if r.Priority < 1 {
		r.Priority = 1
	}
	if r.Priority > 5 {
		r.Priority = 5
	}
	return r
}

func (a Assignment) IsComplete() bool {
	return a.WorkOrderID != "" && a.CrewID != "" && a.Dispatcher != ""
}

func (u RepairUpdate) IsComplete() bool {
	return u.WorkOrderID != "" && u.Technician != "" && u.Action != "" && u.Hours > 0
}

func (d AcceptanceDecision) IsComplete() bool { return d.WorkOrderID != "" && d.Supervisor != "" }
