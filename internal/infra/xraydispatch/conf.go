package xraydispatch

type logObject struct {
	Access      string `json:"access"`
	Error       string `json:"error"`
	Loglevel    string `json:"loglevel"`
	DNSLog      bool   `json:"dnsLog"`
	MaskAddress string `json:"maskAddress"`
}

func defineLevel(l string) string {
	levels := []string{"debug", "info", "warning", "error"}
	for _, lv := range levels {
		if lv == l {
			return l
		}
	}
	return "info"
}

func initLogging(level string) *logObject {
	return &logObject{
		Access:      "",
		Error:       "",
		Loglevel:    level,
		DNSLog:      false,
		MaskAddress: "",
	}
}

type statsObject struct{}

func initStats() *statsObject {
	return &statsObject{}
}

type apiObject struct {
	Tag      string   `json:"tag"`
	Listen   string   `json:"listen"`
	Services []string `json:"services"`
}

func initApiObject() *apiObject {
	return &apiObject{
		Tag:    "api",
		Listen: "127.0.0.1:3000",
		Services: []string{
			"HandlerService",
			"LoggerService",
			"StatsService",
			"RoutingService",
		},
	}
}
