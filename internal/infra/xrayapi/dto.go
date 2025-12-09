package xrayapi

type ClientTraffic struct {
	Email string `json:"email"`
	TX    uint64 `json:"tx"`
	RX    uint64 `json:"rx"`
}

type Traffic struct {
	Type string `json:"type"`
	Tag  string `json:"tag"`
	TX   uint64 `json:"tx"`
	RX   uint64 `json:"rx"`
}

func (t Traffic) IsInbound() bool {
	return t.Type == "inbound"
}

func (t Traffic) IsOutbound() bool {
	return t.Type == "outbound"
}
