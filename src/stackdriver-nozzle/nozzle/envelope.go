package nozzle

import (
	"github.com/cloudfoundry-community/firehose-to-syslog/utils"
	"github.com/cloudfoundry/sonde-go/events"
)

type Envelope struct {
	*events.Envelope
}

func (e *Envelope) getApplicationId() string {
	if e.GetEventType() == events.Envelope_HttpStartStop {
		return utils.FormatUUID(e.GetHttpStartStop().GetApplicationId())
	} else if e.GetEventType() == events.Envelope_LogMessage {
		return e.GetLogMessage().GetAppId()
	} else if e.GetEventType() == events.Envelope_ContainerMetric {
		return e.GetContainerMetric().GetApplicationId()
	} else {
		return ""
	}
}

func (e *Envelope) Labels() map[string]string {
	labels := map[string]string{}

	if e.Origin != nil {
		labels["origin"] = e.GetOrigin()
	}

	if e.EventType != nil {
		labels["event_type"] = e.GetEventType().String()
	}

	if e.Deployment != nil {
		labels["deployment"] = e.GetDeployment()
	}

	if e.Job != nil {
		labels["job"] = e.GetJob()
	}

	if e.Index != nil {
		labels["index"] = e.GetIndex()
	}

	if e.Ip != nil {
		labels["ip"] = e.GetIp()
	}

	if appId := e.getApplicationId(); appId != "" {
		labels["application_id"] = appId
	}

	return labels
}
