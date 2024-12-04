//  SPDX-FileCopyrightText: 2024-2024 OOMOL, Inc. <https://www.oomol.com>
//  SPDX-License-Identifier: MPL-2.0

package config

func NewSlice(values []string) Slice {
	return Slice{Values: values}
}

type Slice struct { // A "mixed-type array" in TOML.
	// Note that the fields below _must_ be exported.  Otherwise the TOML
	// encoder would fail during type reflection.
	Values     []string
	Attributes struct { // Using a struct allows for adding more attributes in the future.
		Append *bool // Nil if not set by the user
	}
}

// Get returns the Slice values or an empty string slice.
func (a *Slice) Get() []string {
	if a.Values == nil {
		return []string{}
	}
	return a.Values
}

// Set overrides the values of the Slice.
func (a *Slice) Set(values []string) {
	a.Values = values
}
