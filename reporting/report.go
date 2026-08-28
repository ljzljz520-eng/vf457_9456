package reporting

import (
	"fmt"
	"sort"
	"streetlight/domain"
	"streetlight/operations"
	"strings"
)

type Report struct {
	Dashboard domain.Dashboard
	Orders    []domain.WorkOrder
}

func Build(service *operations.Service, filter domain.OrderFilter) (Report, error) {
	dashboard, err := service.BuildDashboard()
	if err != nil {
		return Report{}, err
	}
	orders, err := service.ListOrders(filter)
	if err != nil {
		return Report{}, err
	}
	sort.SliceStable(orders, func(i, j int) bool { return orders[i].Status < orders[j].Status })
	return Report{Dashboard: dashboard, Orders: orders}, nil
}

func Render(report Report) string {
	var b strings.Builder
	fmt.Fprintf(&b, "poles=%d active=%d open=%d completed=%d accepted=%d rejected=%d faulted_circuits=%d crews=%d\n", report.Dashboard.TotalPoles, report.Dashboard.ActivePoles, report.Dashboard.OpenOrders, report.Dashboard.CompletedOrders, report.Dashboard.AcceptedOrders, report.Dashboard.RejectedOrders, report.Dashboard.FaultedCircuits, report.Dashboard.AvailableCrews)
	for _, order := range report.Orders {
		fmt.Fprintf(&b, "%s|pole=%s|circuit=%s|priority=%d|status=%s|crew=%s\n", order.ID, order.PoleID, order.CircuitID, order.Priority, order.Status, order.AssignedCrew)
	}
	return b.String()
}

func GroupByStatus(orders []domain.WorkOrder) map[domain.Status]int {
	result := make(map[domain.Status]int)
	for _, order := range orders {
		result[order.Status]++
	}
	return result
}
