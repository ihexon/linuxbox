//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package machine

type APIForwardingState int

var (
	ForwarderBinaryName = "gvproxy"
)

const (
	NoForwarding APIForwardingState = iota
	InForwarding
)

type RemoveOptions struct {
	Force        bool
	SaveImage    bool
	SaveIgnition bool
}
type ResetOptions struct {
	Force bool
}
