package reporting

import (
	"fmt"
	"sort"
	"streetlight/domain"
	"strings"
)

type PrioritySummary struct {
	Band      domain.PriorityBand
	Count     int
	OpenCount int
}

type CircuitSummary struct {
	CircuitID       string
	OrderCount      int
	OpenCount       int
	HighestPriority int
	LatestStatus    domain.Status
}

type TransitionMetric struct {
	Status domain.Status
	Count  int
}

func StatusCounts(orders []domain.WorkOrder) []TransitionMetric {
	counts := make(map[domain.Status]int)
	for _, order := range orders {
		counts[order.Status]++
	}
	result := make([]TransitionMetric, 0, len(counts))
	for status, count := range counts {
		result = append(result, TransitionMetric{Status: status, Count: count})
	}
	sort.SliceStable(result, func(i, j int) bool {
		if domain.StatusRank(result[i].Status) == domain.StatusRank(result[j].Status) {
			return result[i].Status < result[j].Status
		}
		return domain.StatusRank(result[i].Status) < domain.StatusRank(result[j].Status)
	})
	return result
}

func PrioritySummaries(orders []domain.WorkOrder) []PrioritySummary {
	groups := map[domain.PriorityBand]*PrioritySummary{}
	for _, order := range orders {
		band := domain.PriorityBandFor(order.Priority)
		item := groups[band]
		if item == nil {
			item = &PrioritySummary{Band: band}
			groups[band] = item
		}
		item.Count++
		if order.IsOpen() {
			item.OpenCount++
		}
	}
	result := make([]PrioritySummary, 0, len(groups))
	for _, item := range groups {
		result = append(result, *item)
	}
	sort.SliceStable(result, func(i, j int) bool { return priorityBandRank(result[i].Band) > priorityBandRank(result[j].Band) })
	return result
}

func priorityBandRank(band domain.PriorityBand) int {
	switch band {
	case domain.PriorityUrgent:
		return 3
	case domain.PriorityPlanned:
		return 2
	default:
		return 1
	}
}

func CircuitSummaries(orders []domain.WorkOrder) []CircuitSummary {
	groups := make(map[string]*CircuitSummary)
	for _, order := range orders {
		item := groups[order.CircuitID]
		if item == nil {
			item = &CircuitSummary{CircuitID: order.CircuitID, HighestPriority: order.Priority, LatestStatus: order.Status}
			groups[order.CircuitID] = item
		}
		item.OrderCount++
		if order.IsOpen() {
			item.OpenCount++
		}
		if order.Priority > item.HighestPriority {
			item.HighestPriority = order.Priority
		}
		if domain.CompareStatusProgress(order.Status, item.LatestStatus) > 0 {
			item.LatestStatus = order.Status
		}
	}
	result := make([]CircuitSummary, 0, len(groups))
	for _, item := range groups {
		result = append(result, *item)
	}
	sort.SliceStable(result, func(i, j int) bool { return result[i].CircuitID < result[j].CircuitID })
	return result
}

func CompletionRate(orders []domain.WorkOrder) int {
	if len(orders) == 0 {
		return 0
	}
	completed := 0
	for _, order := range orders {
		if order.Status == domain.StatusCompleted || order.Status == domain.StatusAccepted {
			completed++
		}
	}
	return completed * 100 / len(orders)
}

func TransitionMetrics(orders []domain.WorkOrder) []TransitionMetric {
	counts := make(map[domain.Status]int)
	for _, order := range orders {
		for _, transition := range order.History {
			counts[transition.To]++
		}
	}
	result := make([]TransitionMetric, 0, len(counts))
	for status, count := range counts {
		result = append(result, TransitionMetric{Status: status, Count: count})
	}
	sort.SliceStable(result, func(i, j int) bool { return domain.StatusRank(result[i].Status) < domain.StatusRank(result[j].Status) })
	return result
}

func RenderCSV(report Report) string {
	var b strings.Builder
	b.WriteString("order_id,pole_id,circuit_id,priority,status,crew\n")
	for _, order := range report.Orders {
		fmt.Fprintf(&b, "%s,%s,%s,%d,%s,%s\n", order.ID, order.PoleID, order.CircuitID, order.Priority, order.Status, order.AssignedCrew)
	}
	return b.String()
}

func RenderStatusTable(orders []domain.WorkOrder) string {
	lines := []string{"ORDER STATUS PRIORITY CREW"}
	items := append([]domain.WorkOrder(nil), orders...)
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Status == items[j].Status {
			return items[i].ID < items[j].ID
		}
		return domain.StatusRank(items[i].Status) > domain.StatusRank(items[j].Status)
	})
	for _, order := range items {
		lines = append(lines, fmt.Sprintf("%s %s %d %s", order.ID, order.Status, order.Priority, order.AssignedCrew))
	}
	return strings.Join(lines, "\n")
}

func FilterText(orders []domain.WorkOrder, query string) []domain.WorkOrder {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return append([]domain.WorkOrder(nil), orders...)
	}
	result := make([]domain.WorkOrder, 0)
	for _, order := range orders {
		text := strings.ToLower(order.ID + " " + order.PoleID + " " + order.CircuitID + " " + order.Description)
		if strings.Contains(text, query) {
			result = append(result, order)
		}
	}
	return result
}

func GroupOpenByCrew(orders []domain.WorkOrder) map[string][]domain.WorkOrder {
	result := make(map[string][]domain.WorkOrder)
	for _, order := range orders {
		if order.IsOpen() {
			result[order.AssignedCrew] = append(result[order.AssignedCrew], order)
		}
	}
	for crewID := range result {
		sort.SliceStable(result[crewID], func(i, j int) bool { return result[crewID][i].Priority > result[crewID][j].Priority })
	}
	return result
}

func Summarize(report Report) string {
	return fmt.Sprintf("orders=%d completion=%d%% statuses=%d priorities=%d circuits=%d", len(report.Orders), CompletionRate(report.Orders), len(StatusCounts(report.Orders)), len(PrioritySummaries(report.Orders)), len(CircuitSummaries(report.Orders)))
}
