//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package define

type APIForwardingState int

const (
	NoForwarding APIForwardingState = iota
	InForwarding
)
