package domain

import "sort"

type AssetCondition string

const (
	ConditionOperational AssetCondition = "operational"
	ConditionDegraded    AssetCondition = "degraded"
	ConditionOutage      AssetCondition = "outage"
)

type PriorityBand string

const (
	PriorityRoutine PriorityBand = "routine"
	PriorityPlanned PriorityBand = "planned"
	PriorityUrgent  PriorityBand = "urgent"
)

func NormalizeDescription(value string) string {
	result := make([]rune, 0, len(value))
	space := false
	for _, r := range value {
		if r == ' ' || r == '\t' || r == '\n' {
			if len(result) > 0 {
				space = true
			}
			continue
		}
		if space {
			result = append(result, ' ')
			space = false
		}
		result = append(result, r)
	}
	return string(result)
}

func PriorityBandFor(priority int) PriorityBand {
	if priority >= 5 {
		return PriorityUrgent
	}
	if priority >= 3 {
		return PriorityPlanned
	}
	return PriorityRoutine
}

func SeverityForPriority(priority int) int {
	if priority >= 5 {
		return 3
	}
	if priority >= 3 {
		return 2
	}
	return 1
}

func (w WorkOrder) IsOpen() bool {
	return w.Status != StatusAccepted && w.Status != StatusRejected
}

func (w WorkOrder) IsReadyForInspection() bool {
	return w.Status == StatusCompleted || w.Status == StatusDispatched
}

func (w WorkOrder) LastTransition() (StatusTransition, bool) {
	if len(w.History) == 0 {
		return StatusTransition{}, false
	}
	latest := w.History[0]
	for _, item := range w.History[1:] {
		if item.Sequence > latest.Sequence {
			latest = item
		}
	}
	return latest, true
}

func (w WorkOrder) TransitionCount(to Status) int {
	count := 0
	for _, item := range w.History {
		if item.To == to {
			count++
		}
	}
	return count
}

func (w WorkOrder) HasActor(actor string) bool {
	if actor == "" {
		return false
	}
	for _, item := range w.History {
		if item.Actor == actor {
			return true
		}
	}
	return false
}

func (w WorkOrder) ValidateHistory() bool {
	if len(w.History) == 0 {
		return false
	}
	previous := Status("")
	sequence := 0
	for index, item := range w.History {
		if index == 0 {
			if item.From != "" || item.To == "" {
				return false
			}
		} else if item.From != previous || item.Sequence <= sequence || !CanTransition(item.From, item.To) {
			return false
		}
		previous = item.To
		sequence = item.Sequence
	}
	return previous == w.Status && sequence == w.Version
}

func AllowedNextStatuses(from Status) []Status {
	result := make([]Status, 0, 2)
	for _, candidate := range []Status{StatusDispatched, StatusInProgress, StatusCompleted, StatusAccepted, StatusRejected} {
		if CanTransition(from, candidate) {
			result = append(result, candidate)
		}
	}
	return result
}

func SortTransitions(items []StatusTransition) {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Sequence == items[j].Sequence {
			return items[i].Actor < items[j].Actor
		}
		return items[i].Sequence < items[j].Sequence
	})
}

func AssetConditionForPole(pole StreetlightPole, boxes []ControlBox) AssetCondition {
	if !pole.Active {
		return ConditionOutage
	}
	for _, box := range boxes {
		if box.PoleID == pole.ID && !box.Operational {
			return ConditionDegraded
		}
	}
	return ConditionOperational
}

func AssetConditionForCircuit(circuit CircuitLine) AssetCondition {
	if circuit.Faulted {
		return ConditionOutage
	}
	if circuit.Voltage < 200 {
		return ConditionDegraded
	}
	return ConditionOperational
}

func StatusRank(status Status) int {
	switch status {
	case StatusReported:
		return 1
	case StatusDispatched:
		return 2
	case StatusInProgress:
		return 3
	case StatusCompleted:
		return 4
	case StatusAccepted:
		return 5
	case StatusRejected:
		return 0
	default:
		return -1
	}
}

func CompareStatusProgress(left, right Status) int {
	l, r := StatusRank(left), StatusRank(right)
	if l < r {
		return -1
	}
	if l > r {
		return 1
	}
	return 0
}
