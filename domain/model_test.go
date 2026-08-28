package domain

import "testing"

func TestStatusRules(t *testing.T) {
	if !CanTransition(StatusReported, StatusDispatched) || !CanTransition(StatusCompleted, StatusAccepted) {
		t.Fatal("expected legal transitions")
	}
	if CanTransition(StatusAccepted, StatusDispatched) {
		t.Fatal("accepted order must be terminal")
	}
	if !(OrderFilter{Status: StatusAccepted}).Matches(WorkOrder{Status: StatusAccepted}) {
		t.Fatal("filter mismatch")
	}
}
