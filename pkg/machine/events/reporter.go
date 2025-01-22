//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package events

import (
	"net/url"

	"bauklotze/pkg/httpclient"
	allFlag "bauklotze/pkg/machine/allflag"

	"github.com/sirupsen/logrus"
)

// notify sends an event to the report URL
func notify(e event) {
	if allFlag.ReportURL == "" {
		return
	}

	client := httpclient.New().
		SetTransport(httpclient.CreateUnixTransport(allFlag.ReportURL)).
		SetBaseURL("http://local").
		SetHeader("Content-Type", PlainTextContentType).
		SetQueryParams(map[string]string{
			"stage": e.Stage,
			"name":  e.Name,
			"value": url.QueryEscape(e.Value),
		})

	logrus.Infof("Send Event to %s , stage: %s, name: %s, value: %s \n",
		allFlag.ReportURL,
		client.QueryParam.Get("stage"),
		client.QueryParam.Get("name"),
		client.QueryParam.Get("value"),
	)

	if err := client.Get("notify"); err != nil {
		logrus.Warnf("Failed to notify %q: %v\n", allFlag.ReportURL, err)
	}
}

// NotifyInit Generic Notifier for InitStage
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

// NotifyRun Generic Notifier for RunStage
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

// NotifyExit Generic Notifier for Exit
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

// NotifyError Generic Notifier for Error
func NotifyError(err error) {
	notify(event{
		Stage: CurrentStage,
		Name:  kError,
		Value: err.Error(),
	})
}
