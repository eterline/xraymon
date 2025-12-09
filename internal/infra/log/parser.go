package log

import (
	"regexp"
	"strings"
)

var (
	accessReg = regexp.MustCompile(
		`from (?P<client>[^ ]+)\s+accepted\s+(?P<target>[^ ]+)\s+\` +
			`[(?P<inbound>[^ ]+)\s*(?:>>|=>|->)\s*(?P<outbound>[^ ]+)\]` +
			`(?:\s+email:\s*(?P<email>[A-Za-z0-9._\-]+))?`,
	)

	coreLineReg = regexp.MustCompile(`\[[A-Za-z]+\]\s+(.+)$`)
)

type accessFields struct {
	Client   string
	Target   string
	Inbound  string
	Outbound string
	Email    string
}

func (af *accessFields) accessProto() (target, proto string) {
	switch {
	case strings.HasPrefix(af.Target, "udp"):
		return af.Target, "udp"
	case strings.HasPrefix(af.Target, "tcp"):
		return af.Target, "tcp"
	default:
		return af.Target, "http"
	}
}

func (af *accessFields) getEmail() string {
	if af.Email == "" {
		return "UNKNOWN"
	}
	return af.Email
}

func parseAccess(line string) (*accessFields, bool) {
	match := accessReg.FindStringSubmatch(line)
	if match == nil {
		return nil, false
	}

	res := &accessFields{}
	names := accessReg.SubexpNames()

	for i, name := range names {
		if i == 0 || name == "" {
			continue
		}
		switch name {
		case "client":
			res.Client = match[i]
		case "target":
			res.Target = match[i]
		case "inbound":
			res.Inbound = match[i]
		case "outbound":
			res.Outbound = match[i]
		case "email":
			res.Email = match[i]
		}
	}

	return res, true
}

type coreLineFields struct {
	IsError bool
	Payload string
}

func parseCoreLine(p []byte) (*coreLineFields, bool) {
	match := coreLineReg.FindSubmatch(p)
	if len(match) != 2 {
		return nil, false
	}

	var isErr bool

	if len(match[0]) > 0 {
		isErr = match[0][0] == 'E'
	}

	flds := &coreLineFields{
		IsError: isErr,
		Payload: string(match[1]),
	}

	return flds, true
}
