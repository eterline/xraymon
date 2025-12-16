// Copyright (c) 2025 EterLine (Andrew)
// This file is part of xraymon.
// Licensed under the MIT License. See the LICENSE file for details.
package xraymon

type InitFlags struct {
	CommitHash string
	Version    string
	Repository string
}

func (inf InitFlags) GetCommitHash() string {
	return inf.CommitHash
}

func (inf InitFlags) GetVersion() string {
	return inf.Version
}

func (inf InitFlags) GetRepository() string {
	return inf.Repository
}
