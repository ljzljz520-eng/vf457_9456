package dispatch

import (
	"streetlight/domain"
	"testing"
)

func TestRoutePlanUsesReachableStops(t *testing.T) {
	orders := []domain.WorkOrder{{ID: "o1", PoleID: "p2", Priority: 4, AssignedCrew: "crew", Status: domain.StatusReported}, {ID: "o2", PoleID: "p3", Priority: 2, AssignedCrew: "crew", Status: domain.StatusReported}}
	segments := []RoadSegment{{ID: "a", FromPole: "p1", ToPole: "p2", Minutes: 5}, {ID: "b", FromPole: "p2", ToPole: "p3", Minutes: 7}}
	plan, err := PlanRoute("crew", "p1", orders, segments)
	if err != nil || !plan.Feasible() || plan.TotalMinutes != 12 {
		t.Fatalf("plan=%+v err=%v", plan, err)
	}
}
