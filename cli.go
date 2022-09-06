package main

import (
	"github.com/urfave/cli/v2"
)

var (
	logFormatFlag = cli.StringFlag{
		Name:        `log-format`,
		Required:    false,
		Usage:       `Log Format: json, logfmt`,
		Value:       `logfmt`,
		DefaultText: `logfmt`,
		HasBeenSet:  true,
		Category:    `logs`,
		EnvVars: []string{
			`LOG_FORMAT`,
			`GIT_SYNC_LOG_FORMAT`,
		},
	}

	logLevelFlag = cli.StringFlag{
		Name:        `log-level`,
		Required:    false,
		Usage:       `Log Level: panic, fatal, error, warn, info, debug, trace`,
		Value:       `info`,
		DefaultText: `info`,
		HasBeenSet:  true,
		Category:    `logs`,
		EnvVars: []string{
			`LOG_LEVEL`,
			`GIT_SYNC_LOG_LEVEL`,
		},
	}

	logPrettyFlag = cli.BoolFlag{
		Name:        `log-pretty`,
		Required:    false,
		Usage:       `Pretty Log Format (for "json" format)`,
		Value:       false,
		DefaultText: `false`,
		HasBeenSet:  false,
		Category:    `logs`,
		EnvVars: []string{
			`LOG_PRETTY`,
			`GIT_SYNC_LOG_PRETTY`,
		},
	}

	logColorsFlag = cli.BoolFlag{
		Name:        `log-colors`,
		Required:    false,
		Usage:       `Log Colors (for "logfmt" format)`,
		Value:       false,
		DefaultText: `false`,
		HasBeenSet:  false,
		Category:    `logs`,
		EnvVars: []string{
			`LOG_COLORS`,
			`GIT_SYNC_LOG_COLORS`,
		},
	}

	configFileFlag = cli.StringFlag{
		Name:       `config`,
		Required:   true,
		Usage:      `path to the yaml config file`,
		Category:   `config`,
		Value:      ``,
		HasBeenSet: true,
		EnvVars: []string{
			`CONFIG`,
			`GIT_SYNC_CONFIG`,
		},
	}

	startServerFlag = cli.BoolFlag{
		Name:        `server`,
		Required:    false,
		Usage:       `Start HTTP Webserver`,
		Value:       false,
		DefaultText: `false`,
		HasBeenSet:  true,
		Category:    `metrics`,
		EnvVars: []string{
			`GIT_SYNC_SERVER`,
		},
	}

	serverPortFlag = cli.IntFlag{
		Name:        `port`,
		Required:    false,
		Usage:       `HTTP Server Port`,
		Value:       9125,
		DefaultText: `9125`,
		HasBeenSet:  true,
		Category:    `metrics`,
		EnvVars: []string{
			`GIT_SYNC_PORT`,
		},
	}
)
