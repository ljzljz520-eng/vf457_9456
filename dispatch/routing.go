package dispatch

import (
	"fmt"
	"sort"
	"streetlight/domain"
)

type RoadSegment struct {
	ID       string
	FromPole string
	ToPole   string
	Minutes  int
	Blocked  bool
}

type RouteStop struct {
	OrderID  string
	PoleID   string
	Priority int
	Minutes  int
}

type RoutePlan struct {
	CrewID       string
	Stops        []RouteStop
	TotalMinutes int
	BlockedStops int
}

func BuildRoadMap(segments []RoadSegment) map[string][]RoadSegment {
	result := make(map[string][]RoadSegment)
	for _, segment := range segments {
		if segment.ID == "" || segment.FromPole == "" || segment.ToPole == "" || segment.Minutes <= 0 {
			continue
		}
		result[segment.FromPole] = append(result[segment.FromPole], segment)
		reverse := segment
		reverse.FromPole, reverse.ToPole = segment.ToPole, segment.FromPole
		result[reverse.FromPole] = append(result[reverse.FromPole], reverse)
	}
	for key := range result {
		sort.SliceStable(result[key], func(i, j int) bool {
			if result[key][i].Minutes == result[key][j].Minutes {
				return result[key][i].ToPole < result[key][j].ToPole
			}
			return result[key][i].Minutes < result[key][j].Minutes
		})
	}
	return result
}

func RouteDistance(from, to string, segments []RoadSegment) (int, bool) {
	if from == to {
		return 0, true
	}
	graph := BuildRoadMap(segments)
	type point struct {
		pole string
		cost int
	}
	queue := []point{{pole: from, cost: 0}}
	visited := map[string]bool{from: true}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, segment := range graph[current.pole] {
			if segment.Blocked || visited[segment.ToPole] {
				continue
			}
			if segment.ToPole == to {
				return current.cost + segment.Minutes, true
			}
			visited[segment.ToPole] = true
			queue = append(queue, point{pole: segment.ToPole, cost: current.cost + segment.Minutes})
		}
	}
	return 0, false
}

func PlanRoute(crewID, start string, orders []domain.WorkOrder, segments []RoadSegment) (RoutePlan, error) {
	if crewID == "" || start == "" {
		return RoutePlan{}, fmt.Errorf("crew and start pole are required")
	}
	selected := make([]domain.WorkOrder, 0)
	for _, order := range orders {
		if order.AssignedCrew == crewID && order.IsOpen() {
			selected = append(selected, order)
		}
	}
	sort.SliceStable(selected, func(i, j int) bool {
		if selected[i].Priority == selected[j].Priority {
			return selected[i].ID < selected[j].ID
		}
		return selected[i].Priority > selected[j].Priority
	})
	plan := RoutePlan{CrewID: crewID, Stops: make([]RouteStop, 0, len(selected))}
	current := start
	for _, order := range selected {
		minutes, found := RouteDistance(current, order.PoleID, segments)
		if !found {
			plan.BlockedStops++
			plan.Stops = append(plan.Stops, RouteStop{OrderID: order.ID, PoleID: order.PoleID, Priority: order.Priority, Minutes: -1})
			continue
		}
		plan.TotalMinutes += minutes
		plan.Stops = append(plan.Stops, RouteStop{OrderID: order.ID, PoleID: order.PoleID, Priority: order.Priority, Minutes: minutes})
		current = order.PoleID
	}
	return plan, nil
}

func (p RoutePlan) Feasible() bool {
	return len(p.Stops) > 0 && p.BlockedStops == 0
}

func (p RoutePlan) HighestPriority() int {
	highest := 0
	for _, stop := range p.Stops {
		if stop.Priority > highest {
			highest = stop.Priority
		}
	}
	return highest
}

func (p RoutePlan) Summary() string {
	if len(p.Stops) == 0 {
		return "no stops"
	}
	parts := make([]string, 0, len(p.Stops))
	for _, stop := range p.Stops {
		parts = append(parts, stop.OrderID+"@"+stop.PoleID)
	}
	return fmt.Sprintf("crew=%s minutes=%d blocked=%d stops=%s", p.CrewID, p.TotalMinutes, p.BlockedStops, joinRoute(parts))
}

func joinRoute(items []string) string {
	result := ""
	for index, item := range items {
		if index > 0 {
			result += " -> "
		}
		result += item
	}
	return result
}

func FilterRoutableOrders(orders []domain.WorkOrder, crewID string, minimumPriority int) []domain.WorkOrder {
	result := make([]domain.WorkOrder, 0)
	for _, order := range orders {
		if order.AssignedCrew != crewID || !order.IsOpen() || order.Priority < minimumPriority {
			continue
		}
		result = append(result, order)
	}
	SortByUrgency(result)
	return result
}

func SortByUrgency(orders []domain.WorkOrder) {
	sort.SliceStable(orders, func(i, j int) bool {
		if orders[i].Priority == orders[j].Priority {
			return orders[i].Version > orders[j].Version
		}
		return orders[i].Priority > orders[j].Priority
	})
}
