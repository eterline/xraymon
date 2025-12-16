package config

import (
	"os"
	"path/filepath"

	"github.com/alexflint/go-arg"
)

type (
	Log struct {
		LogLevel string `arg:"--log-level" help:"Logging level: debug|info|warn|error"`
		JSONlog  bool   `arg:"--log-json,-j" help:"Set logs to JSON format"`
	}

	Core struct {
		CoreAccess string `arg:"--core-access" help:"Core access file path"`
		CoreLog    string `arg:"--core-log" help:"Core logging file path"`
		ConfigFile string `arg:"--core-config" help:"Core logging file path"`
	}

	Server struct {
		Listen     string `arg:"--listen,-l" help:"Server listen address"`
		CrtFileSSL string `arg:"--certfile,-c" help:"Server SSL certificate file"`
		KeyFileSSL string `arg:"--keyfile,-k" help:"Server SSL key file"`
	}

	Configuration struct {
		Log
		Server
		Core
	}
)

var (
	parserConfig = arg.Config{
		Program:           selfExec(),
		IgnoreEnv:         false,
		IgnoreDefault:     false,
		StrictSubcommands: true,
	}
)

func ParseArgs(c *Configuration) error {
	p, err := arg.NewParser(parserConfig, c)
	if err != nil {
		return err
	}

	err = p.Parse(os.Args[1:])
	if err == arg.ErrHelp {
		p.WriteHelp(os.Stdout)
		os.Exit(1)
	}
	return err
}

func selfExec() string {
	exePath, err := os.Executable()
	if err != nil {
		return "monita"
	}

	return filepath.Base(exePath)
}
