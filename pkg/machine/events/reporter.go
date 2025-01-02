package events

import (
	"bauklotze/pkg/network"
	"net/http"
	"net/url"

	"github.com/sirupsen/logrus"
)

var (
	ReportURL string
)

// notify sends an event to the report URL
func notify(e event) {
	stage := e.Stage
	name := e.Name
	value := e.Value

	if ReportURL == "" {
		return
	}

	connCtx, err := network.NewConn(ReportURL)
	if err != nil {
		logrus.Errorf("report url not valid, --report-url only support unix:/// proto: %v\n", err)
		return
	}

	connCtx.Headers = http.Header{
		"Content-Type": []string{PlainTextContentType},
	}
	connCtx.URLParameter = url.Values{
		"stage": []string{stage},
		"name":  []string{name},
		"value": []string{url.QueryEscape(value)},
	}
	logrus.Infof("Send Event to %s , stage: %s, name: %s, value: %s \n",
		connCtx.URI,
		connCtx.URLParameter.Get("stage"),
		connCtx.URLParameter.Get("name"),
		connCtx.URLParameter.Get("value"),
	)
	req, err := connCtx.Request("GET", "notify")
	if err != nil {
		logrus.Warnf("Failed to notify %q: %v\n", ReportURL, err)
	} else {
		req.Body.Close()
	}
}

func NotifyInit(name InitStageName, value ...string) {
	v := ""
	if len(value) > 0 {
		v = value[0]
	}

	notify(event{
		Stage: Init,
		Name:  string(name),
		Value: v,
	})
}

func NotifyRun(name RunStageName, value ...string) {
	v := ""
	if len(value) > 0 {
		v = value[0]
	}

	notify(event{
		Stage: Run,
		Name:  string(name),
		Value: v,
	})
}

func NotifyExit() {
	switch CurrentStage {
	case Init:
		NotifyInit(InitExit)
	case Run:
		NotifyRun(RunExit)
	default:
		logrus.Warnf("Unknown stage %q", CurrentStage)
	}
}
