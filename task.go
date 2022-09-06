package main

import (
	"errors"
	"fmt"
	"github.com/go-git/go-git/v5"
	log "github.com/sirupsen/logrus"
	"io/fs"
	"net/url"
	"os"
	. "registry.fozzy.lan/palefat/git-sync-go/errors"
)

type GitSyncTask interface {
	CloneOrAttach() error
	Pull() error
}

type gitSyncTask struct {
	config *TaskConfig
	repo   *git.Repository
}

func NewGitSyncTask(config *TaskConfig) (GitSyncTask, error) {
	return &gitSyncTask{
		config: config,
	}, nil
}

func (task *gitSyncTask) createDir() error {

	gitUrl := task.config.Url
	path := task.config.Path

	_, err := os.Stat(path)

	if err == nil {

		log.WithFields(log.Fields{
			`name`: task.config.Name,
			`url`:  gitUrl,
			`path`: path,
		}).Debug(`directory exists`)

		if task.config.Force {

			log.WithFields(log.Fields{
				`name`: task.config.Name,
				`url`:  gitUrl,
				`path`: path,
			}).Debug(`remove directory`)

			err = os.RemoveAll(path)
			if err != nil {

				log.WithError(err).WithFields(log.Fields{
					`name`: task.config.Name,
					`url`:  gitUrl,
					`path`: path,
				}).Error(`unable to remove directory`)

				return err
			}

			err = os.ErrNotExist

		} else {
			return nil
		}

	}

	if errors.Is(err, os.ErrNotExist) {
		err = os.MkdirAll(path, fs.ModePerm)
		if err != nil {

			log.WithError(err).WithFields(log.Fields{
				`name`: task.config.Name,
				`url`:  gitUrl,
				`path`: path,
			}).Error(`cannot create directory for the git repo`)

			return err
		}
	}

	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			`name`: task.config.Name,
			`url`:  gitUrl,
			`path`: path,
		}).Error(`cannot get status for the directory`)
	}

	return err
}

func (task *gitSyncTask) attach(cloneOpts *git.CloneOptions) (*git.Repository, error) {
	repo, err := git.PlainOpen(task.config.Path)
	targetRef := cloneOpts.ReferenceName
	if err != nil {

		errMsg := `unable to attach to the git repo`

		log.WithError(err).WithFields(log.Fields{
			`name`:       task.config.Name,
			`url`:        task.config.Url,
			`path`:       task.config.Path,
			`target_ref`: targetRef,
		}).Error(errMsg)

		return nil, fmt.Errorf(errMsg)
	}

	head, err := repo.Head()
	if err != nil {
		return nil, err
	}

	if len(targetRef) > 0 {
		localRef := head.Name()
		if targetRef != localRef {
			errMsg := `local reference and target reference are different`
			log.WithError(err).WithFields(log.Fields{
				`name`:       task.config.Name,
				`url`:        task.config.Url,
				`path`:       task.config.Path,
				`local_ref`:  localRef,
				`target_ref`: targetRef,
			}).Error(errMsg)
			return nil, errors.New(errMsg)
		}

	}

	// TODO: validate url from config and from the repo on the disk
	// TODO: validate revision from config and from the repo on the disk

	return repo, nil
}

func (task *gitSyncTask) doClone(cloneOpts *git.CloneOptions) (*git.Repository, error) {
	repo, err := git.PlainClone(task.config.Path, false, cloneOpts)
	if err == nil || err == git.ErrRepositoryAlreadyExists {
		return repo, err
	}

	log.WithError(err).WithFields(log.Fields{
		`name`: task.config.Name,
		`url`:  task.config.Url,
		`path`: task.config.Path,
	}).Warn(`unable to clone repo by go-git api, run git-clone manually`)

	// FIXME: go-git cannot clone a git repository from Azure DevOps
	// manual clone

	cmd, err := task.config.GitCloneCmd()
	if err != nil {
		return nil, err
	}

	err = cmd.Run()
	if err != nil {
		return nil, err
	}

	return nil, git.ErrRepositoryAlreadyExists
}

func (task *gitSyncTask) CloneOrAttach() error {

	if len(task.config.Url) == 0 {
		return ErrGitRepoUrlIsMissing
	}

	parsedUrl, err := url.Parse(task.config.Url)
	if err != nil {
		return ErrGitRepoUrlIsNotValid
	}

	if parsedUrl.Scheme != `http` && parsedUrl.Scheme != `https` {
		return ErrGitRepoUrlSchemaIsNotSupported
	}

	if len(task.config.Path) == 0 {
		return fmt.Errorf(`GitSyncTask: Target directory (path) is not set`)
	}

	if err = task.createDir(); err != nil {
		return err
	}

	// TODO: check if directory contains files
	// TODO: remove files in the directory is flag force is set

	cloneOpts, err := task.config.CloneOptions()
	if err != nil {
		return err
	}

	repo, err := task.doClone(cloneOpts)
	if err == git.ErrRepositoryAlreadyExists {

		repo, err = task.attach(cloneOpts)
		if err != nil {
			return err
		}

	} else if err != nil {

		errMsg := `unable to clone the repo`

		log.WithError(err).WithFields(log.Fields{
			`name`:       task.config.Name,
			`url`:        task.config.Url,
			`path`:       task.config.Path,
			`target_ref`: cloneOpts.ReferenceName,
		}).Error(errMsg)

		return fmt.Errorf(errMsg)
	}

	head, err := repo.Head()
	if err != nil {
		return err
	}

	if log.IsLevelEnabled(log.DebugLevel) {

		log.WithFields(log.Fields{
			`name`:       task.config.Name,
			`url`:        task.config.Url,
			`path`:       task.config.Path,
			`target_ref`: cloneOpts.ReferenceName,
			`local_ref`:  head.Name(),
		}).Debug(`repo has been cloned`)
	}

	task.repo = repo

	return err
}

func (task *gitSyncTask) Pull() error {
	repo := task.repo

	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}

	pullOptions, err := task.config.PullOptions()
	if err != nil {
		return err
	}

	err = worktree.Reset(&git.ResetOptions{
		Mode: git.HardReset,
	})
	if err != nil {
		return err
	}

	err = worktree.Pull(pullOptions)
	if err == git.NoErrAlreadyUpToDate {

		log.WithFields(log.Fields{
			`name`:       task.config.Name,
			`url`:        task.config.Url,
			`path`:       task.config.Path,
			`target_ref`: pullOptions.ReferenceName,
		}).Debug(`repo is up to date`)

		return nil
	}

	return err
}
