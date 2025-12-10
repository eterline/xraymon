// Copyright (c) 2025 EterLine (Andrew)
// This file is part of xraymon.
// Licensed under the MIT License. See the LICENSE file for details.
package domain

type ConnectionMetadata struct {
	Client   string
	Server   string
	Proto    string
	Inbound  string
	Outbound string
	User     string
}
