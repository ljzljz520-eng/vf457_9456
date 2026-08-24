package inspection

import (
	"sort"
	"streetlight/domain"
)

func Latest(items []domain.Inspection, orderID string) (domain.Inspection, bool) {
	filtered := make([]domain.Inspection, 0)
	for _, item := range items {
		if item.WorkOrderID == orderID {
			filtered = append(filtered, item)
		}
	}
	if len(filtered) == 0 {
		return domain.Inspection{}, false
	}
	sort.Slice(filtered, func(i, j int) bool { return filtered[i].Sequence > filtered[j].Sequence })
	return filtered[0], true
}
