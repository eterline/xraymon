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

func NewDataIO() *DataIO {
	return &DataIO{
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
	(*io) = *NewDataIO()
}

type MetricType string

const (
	TypeUser     MetricType = "user"
	TypeInbound  MetricType = "inbound"
	TypeOubnound MetricType = "user"
)

type MetricSnapshot struct {
	Type MetricType
	Name string
	IO   *DataIO
}

func NewUserMetric(email string) *MetricSnapshot {
	return &MetricSnapshot{
		Type: TypeUser,
		Name: email,
		IO:   NewDataIO(),
	}
}

func NewInboundMetric(tag string) *MetricSnapshot {
	return &MetricSnapshot{
		Type: TypeInbound,
		Name: tag,
		IO:   NewDataIO(),
	}
}

func NewOutboundMetric(tag string) *MetricSnapshot {
	return &MetricSnapshot{
		Type: TypeOubnound,
		Name: tag,
		IO:   NewDataIO(),
	}
}
