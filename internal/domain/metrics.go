// Copyright (c) 2025 EterLine (Andrew)
// This file is part of xraymon.
// Licensed under the MIT License. See the LICENSE file for details.
package domain

import "time"

type DataIO struct {
	lastUpdateRx int64
	lastUpdateTX int64

	RX uint64
	TX uint64

	PerSecRX uint64
	PerSecTX uint64
}

func NewDataIO() DataIO {
	return DataIO{
		lastUpdateRx: time.Now().UnixMilli(),
		lastUpdateTX: time.Now().UnixMilli(),
	}
}

func (io *DataIO) deltaRX() float64 {
	is := time.Now()
	last := time.UnixMilli(io.lastUpdateRx)
	io.lastUpdateRx = is.UnixMilli()
	return is.Sub(last).Seconds()
}

func (io *DataIO) deltaTX() float64 {
	is := time.Now()
	last := time.UnixMilli(io.lastUpdateTX)
	io.lastUpdateTX = is.UnixMilli()
	return is.Sub(last).Seconds()
}

func (io *DataIO) IncRX(v uint64) {
	sec := io.deltaRX()
	io.RX += v
	io.PerSecRX = uint64(float64(v) / sec)
}

func (io *DataIO) IncTX(v uint64) {
	sec := io.deltaTX()
	io.TX += v
	io.PerSecTX = uint64(float64(v) / sec)
}

func (io *DataIO) Reset() {
	(*io) = NewDataIO()
}

type StatsType string

const (
	TypeUser     StatsType = "user"
	TypeInbound  StatsType = "inbound"
	TypeOubnound StatsType = "user"
)

type StatsSnapshot struct {
	Type StatsType
	Name string
	IO   DataIO
}

func NewUserMetric(email string) StatsSnapshot {
	return StatsSnapshot{
		Type: TypeUser,
		Name: email,
		IO:   NewDataIO(),
	}
}

func NewInboundMetric(tag string) StatsSnapshot {
	return StatsSnapshot{
		Type: TypeInbound,
		Name: tag,
		IO:   NewDataIO(),
	}
}

func NewOutboundMetric(tag string) StatsSnapshot {
	return StatsSnapshot{
		Type: TypeOubnound,
		Name: tag,
		IO:   NewDataIO(),
	}
}
