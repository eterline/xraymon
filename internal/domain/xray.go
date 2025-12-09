package domain

import "encoding/json"

type CoreConfiguration map[string]json.RawMessage

type ConfigLoader interface {
	Load() (CoreConfiguration, error)
}
