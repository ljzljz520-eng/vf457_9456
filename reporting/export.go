package reporting

import (
	"encoding/json"
	"fmt"
	"streetlight/domain"
)

func JSON(report Report) ([]byte, error) { return json.MarshalIndent(report, "", "  ") }

func StatusLabel(status domain.Status) string {
	switch status {
	case domain.StatusReported:
		return "Reported"
	case domain.StatusDispatched:
		return "Dispatched"
	case domain.StatusInProgress:
		return "In progress"
	case domain.StatusCompleted:
		return "Completed"
	case domain.StatusAccepted:
		return "Accepted"
	case domain.StatusRejected:
		return "Rejected"
	default:
		return fmt.Sprintf("Unknown(%s)", status)
	}
}

func PriorityLabel(priority int) string {
	if priority >= 5 {
		return "urgent"
	}
	if priority >= 3 {
		return "high"
	}
	return "normal"
}
