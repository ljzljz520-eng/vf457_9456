package inspection

import (
	"fmt"
	"sort"
	"streetlight/domain"
)

type ComplianceRule struct {
	Code        string
	Description string
	Required    bool
	Severity    int
}

type ComplianceResult struct {
	OrderID       string
	Passed        bool
	RequiredTotal int
	Satisfied     int
	MissingCodes  []string
	Summary       string
}

var defaultRules = []ComplianceRule{
	{Code: "POWER", Description: "power restored", Required: true, Severity: 5},
	{Code: "LAMP", Description: "lamp aligned", Required: true, Severity: 3},
	{Code: "CABINET", Description: "cabinet locked", Required: true, Severity: 4},
	{Code: "SAFETY", Description: "area safe", Required: true, Severity: 5},
}

func DefaultRules() []ComplianceRule {
	result := make([]ComplianceRule, len(defaultRules))
	copy(result, defaultRules)
	return result
}

func EvaluateChecklist(orderID string, checklist Checklist, rules []ComplianceRule) ComplianceResult {
	if len(rules) == 0 {
		rules = DefaultRules()
	}
	result := ComplianceResult{OrderID: orderID, MissingCodes: make([]string, 0)}
	for _, rule := range rules {
		if !rule.Required {
			continue
		}
		result.RequiredTotal++
		if checklistValue(checklist, rule.Code) {
			result.Satisfied++
		} else {
			result.MissingCodes = append(result.MissingCodes, rule.Code)
		}
	}
	result.Passed = result.RequiredTotal > 0 && result.Satisfied == result.RequiredTotal
	result.Summary = ComplianceSummary(result)
	return result
}

func checklistValue(checklist Checklist, code string) bool {
	switch code {
	case "POWER":
		return checklist.PowerRestored
	case "LAMP":
		return checklist.LampAligned
	case "CABINET":
		return checklist.CabinetLocked
	case "SAFETY":
		return checklist.AreaSafe
	default:
		return false
	}
}

func ComplianceSummary(result ComplianceResult) string {
	if result.Passed {
		return fmt.Sprintf("%s passed %d/%d required checks", result.OrderID, result.Satisfied, result.RequiredTotal)
	}
	return fmt.Sprintf("%s missing %d/%d checks: %s", result.OrderID, len(result.MissingCodes), result.RequiredTotal, joinCodes(result.MissingCodes))
}

func joinCodes(codes []string) string {
	result := ""
	for index, code := range codes {
		if index > 0 {
			result += ","
		}
		result += code
	}
	return result
}

func BuildFindingRegister(findings []Finding) map[string]Finding {
	result := make(map[string]Finding)
	for _, finding := range findings {
		if finding.Code == "" {
			continue
		}
		current, found := result[finding.Code]
		if !found || finding.Severity > current.Severity {
			result[finding.Code] = finding
		}
	}
	return result
}

func CriticalFindings(findings []Finding) []Finding {
	result := make([]Finding, 0)
	for _, finding := range findings {
		if finding.Severity >= 4 {
			result = append(result, finding)
		}
	}
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Severity == result[j].Severity {
			return result[i].Code < result[j].Code
		}
		return result[i].Severity > result[j].Severity
	})
	return result
}

func RequiresReinspection(item domain.Inspection, findings []Finding) bool {
	if item.Passed {
		return false
	}
	for _, finding := range findings {
		if finding.Severity >= 4 {
			return true
		}
	}
	return item.Findings == ""
}

func InspectionOutcome(item domain.Inspection) domain.Status {
	if item.Passed {
		return domain.StatusAccepted
	}
	return domain.StatusRejected
}

func SortInspections(items []domain.Inspection) {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Sequence == items[j].Sequence {
			return items[i].ID < items[j].ID
		}
		return items[i].Sequence > items[j].Sequence
	})
}

func LatestPassed(items []domain.Inspection, orderID string) (domain.Inspection, bool) {
	filtered := make([]domain.Inspection, 0)
	for _, item := range items {
		if item.WorkOrderID == orderID && item.Passed {
			filtered = append(filtered, item)
		}
	}
	SortInspections(filtered)
	if len(filtered) == 0 {
		return domain.Inspection{}, false
	}
	return filtered[0], true
}

func ValidateInspection(item domain.Inspection) error {
	if item.ID == "" || item.WorkOrderID == "" || item.Supervisor == "" {
		return fmt.Errorf("inspection identity is incomplete")
	}
	if item.Sequence <= 0 {
		return fmt.Errorf("inspection sequence must be positive")
	}
	if !item.Passed && item.Findings == "" {
		return fmt.Errorf("failed inspection needs findings")
	}
	return nil
}
