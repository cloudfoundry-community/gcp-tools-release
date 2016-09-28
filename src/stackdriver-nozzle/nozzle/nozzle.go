package nozzle

import (
	"github.com/evandbrown/gcp-tools-boshrelease/src/stackdriver-nozzle/stackdriver"
)


type Nozzle struct {
	StackdriverClient stackdriver.Client
}

func (n *Nozzle) Connect() bool {
	return true
}

func (n *Nozzle) ShipEvents(event map[string]interface{}, _ string /* TODO research second string */) {
	n.StackdriverClient.Post(event, map[string]string{})
}
