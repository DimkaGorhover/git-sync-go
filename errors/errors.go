package errors

import "errors"

var (
	ErrAppIsDone                      = errors.New(`app is finished successfully`)
	ErrGitRepoUrlIsMissing            = errors.New(`git repo url is missing`)
	ErrGitRepoUrlIsNotValid           = errors.New(`git repo url is not valid`)
	ErrGitRepoUrlSchemaIsNotSupported = errors.New(`git repo url schema is not supported`)
)
