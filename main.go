package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v2"
	"os"
	. "registry.fozzy.lan/palefat/git-sync-go/errors"
	. "registry.fozzy.lan/palefat/git-sync-go/scheduler"
)

var (
	AppVersion = "development"
	appErrChan = make(chan error)
	scheduler  = NewAppScheduler(appErrChan)
)

func main() {
	cliApp := &cli.App{
		Name:                 `git-sync`,
		Version:              AppVersion,
		EnableBashCompletion: true,
		ErrWriter:            NewLogrusWriter(log.ErrorLevel),
		Writer:               NewLogrusWriter(log.DebugLevel),
		Authors: []*cli.Author{
			{
				Name:  `Dmytro Horkhover`,
				Email: `gd.mail.89@gmail.com`,
			},
		},
		Flags: []cli.Flag{
			&logLevelFlag,
			&logFormatFlag,
			&logPrettyFlag,
			&logColorsFlag,
			&startServerFlag,
			&serverPortFlag,
		},
		Before: func(c *cli.Context) error {
			return configureLogs(LogConfig{
				level:  c.String(logLevelFlag.Name),
				format: c.String(logFormatFlag.Name),
				pretty: c.Bool(logPrettyFlag.Name),
				colors: c.Bool(logColorsFlag.Name),
			})
		},
		Action: cliAction,
	}

	defer func() {
		_ = scheduler.Close()
		log.Info(`task scheduler has been closed`)
	}()

	if err := cliApp.Run(os.Args); err != nil && err != ErrAppIsDone {
		log.WithError(err).Fatal(`app exit error`)
	}
}

func cliAction(c *cli.Context) error {

	serverPort := c.Int(serverPortFlag.Name)
	if serverPort < 1 {
		return fmt.Errorf(`web server port cannot be less than 1`)
	}

	scheduleTasks()

	if c.Bool(startServerFlag.Name) {
		scheduler.Execute(func() error {
			return startServer(serverPort)
		})
	}

	return scheduler.WaitError()
}
func scheduleTasks() {
	scheduler.Execute(func() error {
		file, err := os.Open("tasks.yaml")
		if err != nil {
			return err
		}

		defer func() {
			_ = file.Close()
		}()

		config := &Config{}
		err = yaml.NewDecoder(file).Decode(config)
		if err != nil {
			return err
		}
		for _, taskConfig := range config.Tasks {
			scheduleTask(taskConfig)
		}
		return nil
	})
}

func scheduleTask(tc *TaskConfig) {
	scheduler.Execute(func() error {
		task, err := NewGitSyncTask(tc)
		if err != nil {
			return err
		}
		if err = task.CloneOrAttach(); err != nil {
			return err
		}
		if tc.RunOnce {
			log.Warn(`run-once mode is not fully supported yet. application will not stop if all task are run-once mode`)
		} else {
			scheduler.Schedule(task.Pull, tc.Interval())
		}
		return nil
	})
}
