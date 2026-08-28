package inspection

import (
	"fmt"
	"streetlight/domain"
	"streetlight/operations"
)

type Checklist struct {
	PowerRestored bool
	LampAligned   bool
	CabinetLocked bool
	AreaSafe      bool
}

func (c Checklist) Complete() bool {
	return c.PowerRestored && c.LampAligned && c.CabinetLocked && c.AreaSafe
}

type Inspector struct{ Service *operations.Service }

func NewInspector(service *operations.Service) *Inspector { return &Inspector{Service: service} }

func (i *Inspector) Review(orderID, inspectionID, supervisor string, checklist Checklist, findings string) (domain.Inspection, error) {
	if supervisor == "" {
		return domain.Inspection{}, fmt.Errorf("supervisor is required")
	}
	passed := checklist.Complete()
	if !passed && findings == "" {
		findings = "checklist incomplete"
	}
	return i.Service.AcceptOrder(domain.AcceptanceDecision{WorkOrderID: orderID, Supervisor: supervisor, Passed: passed, Findings: findings}, inspectionID)
}

func (i *Inspector) Reinspect(orderID, inspectionID, supervisor string, checklist Checklist, findings string) (domain.Inspection, error) {
	if err := i.Service.ReopenRejected(orderID, supervisor, "reinspection requested"); err != nil {
		return domain.Inspection{}, err
	}
	return i.Review(orderID, inspectionID, supervisor, checklist, findings)
}
