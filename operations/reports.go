package operations

import (
	"fmt"
	"sort"
	"streetlight/domain"
	"strings"
)

type RepeatReport struct {
	PoleID       string
	CircuitID    string
	Count        int
	OpenOrderIDs []string
	LatestOrder  string
}

type ReportMatch struct {
	OrderID  string
	Reason   string
	Priority int
}

func NormalizeFaultReport(report domain.FaultReport) (domain.FaultReport, error) {
	report = report.Normalize()
	report.Description = domain.NormalizeDescription(report.Description)
	report.Reporter = strings.TrimSpace(report.Reporter)
	if report.PoleID == "" || report.Reporter == "" || report.Description == "" {
		return domain.FaultReport{}, fmt.Errorf("fault report needs pole, reporter, and description")
	}
	return report, nil
}

func (s *Service) FindRepeatReports(report domain.FaultReport) ([]ReportMatch, error) {
	normalized, err := NormalizeFaultReport(report)
	if err != nil {
		return nil, err
	}
	orders, err := s.DB.ListWorkOrders()
	if err != nil {
		return nil, err
	}
	result := make([]ReportMatch, 0)
	for _, order := range orders {
		if order.PoleID != normalized.PoleID || order.CircuitID != normalized.CircuitID || !order.IsOpen() {
			continue
		}
		if strings.EqualFold(domain.NormalizeDescription(order.Description), normalized.Description) {
			result = append(result, ReportMatch{OrderID: order.ID, Reason: "same asset and description", Priority: order.Priority})
		}
	}
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Priority == result[j].Priority {
			return result[i].OrderID < result[j].OrderID
		}
		return result[i].Priority > result[j].Priority
	})
	return result, nil
}

func (s *Service) ReportFaultWithDuplicateCheck(report domain.FaultReport, orderID string) (domain.WorkOrder, []ReportMatch, error) {
	normalized, err := NormalizeFaultReport(report)
	if err != nil {
		return domain.WorkOrder{}, nil, err
	}
	matches, err := s.FindRepeatReports(normalized)
	if err != nil {
		return domain.WorkOrder{}, nil, err
	}
	order, err := s.ReportFault(normalized, orderID)
	if err != nil {
		return domain.WorkOrder{}, matches, err
	}
	if len(matches) > 0 {
		order.Description = domain.NormalizeDescription(order.Description) + " [repeat of " + matches[0].OrderID + "]"
		order.Version++
		order.History = append(order.History, domain.StatusTransition{Sequence: order.Version, From: order.Status, To: order.Status, Actor: normalized.Reporter, Note: "repeat report linked"})
		if err = s.DB.SaveWorkOrder(order); err != nil {
			return domain.WorkOrder{}, matches, err
		}
	}
	return order, matches, nil
}

func (s *Service) RepeatReportSummary() ([]RepeatReport, error) {
	orders, err := s.DB.ListWorkOrders()
	if err != nil {
		return nil, err
	}
	groups := make(map[string]*RepeatReport)
	for _, order := range orders {
		key := order.PoleID + "|" + order.CircuitID + "|" + domain.NormalizeDescription(order.Description)
		item := groups[key]
		if item == nil {
			item = &RepeatReport{PoleID: order.PoleID, CircuitID: order.CircuitID}
			groups[key] = item
		}
		item.Count++
		if order.IsOpen() {
			item.OpenOrderIDs = append(item.OpenOrderIDs, order.ID)
		}
		if item.LatestOrder == "" || order.Version > versionForOrder(orders, item.LatestOrder) {
			item.LatestOrder = order.ID
		}
	}
	result := make([]RepeatReport, 0, len(groups))
	for _, item := range groups {
		if item.Count > 1 {
			sort.Strings(item.OpenOrderIDs)
			result = append(result, *item)
		}
	}
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Count == result[j].Count {
			return result[i].PoleID < result[j].PoleID
		}
		return result[i].Count > result[j].Count
	})
	return result, nil
}

func versionForOrder(orders []domain.WorkOrder, id string) int {
	for _, order := range orders {
		if order.ID == id {
			return order.Version
		}
	}
	return -1
}

func (s *Service) ReopenOrderForRepeat(orderID, actor, reason string) error {
	if strings.TrimSpace(reason) == "" {
		return fmt.Errorf("repeat reason is required")
	}
	order, err := s.DB.GetWorkOrder(orderID)
	if err != nil {
		return err
	}
	if order.Status != domain.StatusRejected {
		return fmt.Errorf("only rejected orders can be reopened")
	}
	return s.ReopenRejected(orderID, actor, domain.NormalizeDescription(reason))
}

func MatchReportText(order domain.WorkOrder, query string) bool {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return true
	}
	text := strings.ToLower(order.ID + " " + order.PoleID + " " + order.CircuitID + " " + order.Description)
	return strings.Contains(text, query)
}
