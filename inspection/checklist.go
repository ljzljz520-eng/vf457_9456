package inspection

import "streetlight/domain"

type Finding struct {
	Code        string
	Description string
	Severity    int
}

func SummarizeFindings(findings []Finding) string {
	if len(findings) == 0 {
		return "no findings"
	}
	result := ""
	for index, finding := range findings {
		if index > 0 {
			result += "; "
		}
		result += finding.Code + ":" + finding.Description
		if finding.Severity >= 4 {
			result += " (critical)"
		}
	}
	return result
}

func ChecklistFromFindings(findings []Finding) Checklist {
	result := Checklist{PowerRestored: true, LampAligned: true, CabinetLocked: true, AreaSafe: true}
	for _, finding := range findings {
		switch finding.Code {
		case "POWER":
			result.PowerRestored = false
		case "LAMP":
			result.LampAligned = false
		case "CABINET":
			result.CabinetLocked = false
		case "SAFETY":
			result.AreaSafe = false
		}
	}
	return result
}

func InspectionPassed(item domain.Inspection) bool {
	return item.Passed && item.Findings != "" || item.Passed
}
