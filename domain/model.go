package domain

import "fmt"

type Status string

const (
	StatusReported   Status = "reported"
	StatusDispatched Status = "dispatched"
	StatusInProgress Status = "in_progress"
	StatusCompleted  Status = "completed"
	StatusAccepted   Status = "accepted"
	StatusRejected   Status = "rejected"
)

type StreetlightPole struct {
	ID       string `json:"id"`
	Code     string `json:"code"`
	Location string `json:"location"`
	LampType string `json:"lamp_type"`
	Active   bool   `json:"active"`
}

type ControlBox struct {
	ID          string `json:"id"`
	PoleID      string `json:"pole_id"`
	CircuitID   string `json:"circuit_id"`
	Address     string `json:"address"`
	Operational bool   `json:"operational"`
}

type CircuitLine struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Voltage   int      `json:"voltage"`
	PoleIDs   []string `json:"pole_ids"`
	Faulted   bool     `json:"faulted"`
	FaultNote string   `json:"fault_note"`
}

type Crew struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Members []string `json:"members"`
	Skills  []string `json:"skills"`
	Active  bool     `json:"active"`
}

type StatusTransition struct {
	Sequence int    `json:"sequence"`
	From     Status `json:"from"`
	To       Status `json:"to"`
	Actor    string `json:"actor"`
	Note     string `json:"note"`
}

type WorkOrder struct {
	ID           string             `json:"id"`
	PoleID       string             `json:"pole_id"`
	CircuitID    string             `json:"circuit_id"`
	Description  string             `json:"description"`
	Priority     int                `json:"priority"`
	Status       Status             `json:"status"`
	AssignedCrew string             `json:"assigned_crew"`
	Version      int                `json:"version"`
	History      []StatusTransition `json:"history"`
}

type Inspection struct {
	ID          string `json:"id"`
	WorkOrderID string `json:"work_order_id"`
	Supervisor  string `json:"supervisor"`
	Passed      bool   `json:"passed"`
	Findings    string `json:"findings"`
	Sequence    int    `json:"sequence"`
}

type RepairRecord struct {
	ID          string `json:"id"`
	WorkOrderID string `json:"work_order_id"`
	Action      string `json:"action"`
	Material    string `json:"material"`
	Hours       int    `json:"hours"`
	CompletedBy string `json:"completed_by"`
}

func (s Status) Valid() bool {
	switch s {
	case StatusReported, StatusDispatched, StatusInProgress, StatusCompleted, StatusAccepted, StatusRejected:
		return true
	default:
		return false
	}
}

func (s Status) String() string { return string(s) }

func CanTransition(from, to Status) bool {
	if from == StatusReported && to == StatusDispatched {
		return true
	}
	if from == StatusDispatched && to == StatusInProgress {
		return true
	}
	if from == StatusInProgress && to == StatusCompleted {
		return true
	}
	if from == StatusCompleted && to == StatusAccepted {
		return true
	}
	if from == StatusCompleted && to == StatusRejected {
		return true
	}
	if from == StatusRejected && to == StatusInProgress {
		return true
	}
	return false
}

func (w WorkOrder) Validate() error {
	if w.ID == "" {
		return fmt.Errorf("work order id is required")
	}
	if w.PoleID == "" {
		return fmt.Errorf("pole id is required")
	}
	if !w.Status.Valid() {
		return fmt.Errorf("invalid status %q", w.Status)
	}
	if w.Priority < 1 || w.Priority > 5 {
		return fmt.Errorf("priority must be 1..5")
	}
	return nil
}

func (p StreetlightPole) Validate() error {
	if p.ID == "" || p.Code == "" || p.Location == "" {
		return fmt.Errorf("pole identity and location are required")
	}
	return nil
}

func (c ControlBox) Validate() error {
	if c.ID == "" || c.PoleID == "" || c.CircuitID == "" {
		return fmt.Errorf("control box links are required")
	}
	return nil
}

func (c CircuitLine) Validate() error {
	if c.ID == "" || c.Name == "" {
		return fmt.Errorf("circuit identity is required")
	}
	if c.Voltage <= 0 {
		return fmt.Errorf("voltage must be positive")
	}
	return nil
}

func (c Crew) Validate() error {
	if c.ID == "" || c.Name == "" {
		return fmt.Errorf("crew identity is required")
	}
	if len(c.Members) == 0 {
		return fmt.Errorf("crew needs members")
	}
	return nil
}
