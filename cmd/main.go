//  SPDX-FileCopyrightText: 2024-2025 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"
	"os"

	"bauklotze/pkg/machine/events"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v3"
)

func main() {
	app := cli.Command{
		Commands: []*cli.Command{
			&initCmd,
			&startCmd,
		},
	}

	NotifyAndExit(app.Run(context.Background(), os.Args))
}

func NotifyAndExit(err error) {
	retCode := 0
	if err != nil {
		retCode = 1
		logrus.Error(err.Error())
		events.NotifyError(err)
	}
	events.NotifyExit()
	logrus.Exit(retCode)
}
