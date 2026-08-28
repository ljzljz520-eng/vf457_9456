package reporting

import (
	"streetlight/domain"
	"strings"
	"testing"
)

func TestRenderReport(t *testing.T) {
	text := Render(Report{Dashboard: domain.Dashboard{TotalPoles: 2, ActivePoles: 1}, Orders: []domain.WorkOrder{{ID: "o", PoleID: "p", Priority: 5, Status: domain.StatusAccepted}}})
	if !strings.Contains(text, "accepted=0") || !strings.Contains(text, "o|pole=p") {
		t.Fatal(text)
	}
	if StatusLabel(domain.StatusAccepted) != "Accepted" || PriorityLabel(5) != "urgent" {
		t.Fatal("labels mismatch")
	}
}
