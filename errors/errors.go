package errors

import "errors"

var (
	ErrAppIsDone           = errors.New(`app is finished successfully`)
	ErrGitRepoUrlIsMissing = errors.New(`git repo url is missing`)
)
