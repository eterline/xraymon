package xrayapi

import (
	"bytes"

	statsService "github.com/xtls/xray-core/app/stats/command"
)

// ===================================

func isEqBytes(s, v []byte) bool {
	return bytes.Equal(s, v)
}

type io struct {
	up   uint64
	down uint64
}

func (io *io) addUp(v uint64) {
	io.up += v
}

func (io *io) addDown(v uint64) {
	io.down += v
}

// ===================================

func parseStats(s []*statsService.Stat) ([]Traffic, []ClientTraffic) {
	var (
		email        = map[string]*io{}
		tag          = map[string]*io{}
		tagIsInbound = map[string][]byte{}
	)

	for _, stat := range s {

		nameBytes := []byte(stat.Name)

		if matches := clientTrafficReg.FindSubmatch(nameBytes); len(matches) == 3 {
			var (
				emailName = matches[2]
				down      = isEqBytes(matches[2], []byte("downlink"))
			)

			v, ok := email[string(emailName)]
			if !ok {
				v = &io{}
				email[string(emailName)] = v
			}

			if down {
				v.addDown(uint64(stat.Value))
			} else {
				v.addUp(uint64(stat.Value))
			}
		}

		if matches := trafficReg.FindSubmatch(nameBytes); len(matches) == 4 {
			var (
				tagName = matches[2]
				down    = isEqBytes(matches[3], []byte("downlink"))
			)

			if isEqBytes(tagName, []byte("api")) {
				continue
			}

			if len(tagName) == 0 {
				tagName = append(tagName, []byte("NULL")...)
			}

			tagIsInbound[string(tagName)] = matches[1]

			v, ok := tag[string(tagName)]
			if !ok {
				v = &io{}
				tag[string(tagName)] = v
			}

			if down {
				v.addDown(uint64(stat.Value))
			} else {
				v.addUp(uint64(stat.Value))
			}
		}
	}

	var (
		clientTraffic = make([]ClientTraffic, 0, len(email))
		traffic       = make([]Traffic, 0, len(tag))
	)

	for i, s := range email {
		v := ClientTraffic{
			Email: i,
			TX:    s.up,
			RX:    s.down,
		}

		clientTraffic = append(clientTraffic, v)
	}

	for i, s := range tag {
		v := Traffic{
			Type: string(tagIsInbound[i]),
			Tag:  i,
			TX:   s.up,
			RX:   s.down,
		}

		traffic = append(traffic, v)
	}

	return traffic, clientTraffic
}
