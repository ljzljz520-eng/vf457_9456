package inspection

import "testing"

func TestChecklistSummary(t *testing.T) {
	items := []Finding{{Code: "POWER", Description: "breaker open", Severity: 5}, {Code: "LAMP", Description: "tilted", Severity: 2}}
	if got := SummarizeFindings(items); got == "" {
		t.Fatal("summary empty")
	}
	check := ChecklistFromFindings(items)
	if check.Complete() {
		t.Fatal("failed checklist marked complete")
	}
}
