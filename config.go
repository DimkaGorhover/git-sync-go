package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	gitTransport "github.com/go-git/go-git/v5/plumbing/transport"
	gitHttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	log "github.com/sirupsen/logrus"
	"net/url"
	"os"
	"os/exec"
	. "registry.fozzy.lan/palefat/git-sync-go/git"
	"strings"
	"time"
)

const (
	defaultIntervalSeconds = 60
)

var (
	ErrBasicAuthUserIsMissing     = errors.New(`user is not configured for basic config`)
	ErrBasicAuthPasswordIsMissing = errors.New(`password is not configured for basic config`)

	ErrNameIsMissing   = errors.New(`task name is missing`)
	ErrNameIsNotUnique = errors.New(`task name is not unique`)
	ErrPathIsMissing   = errors.New(`task path is missing`)
	ErrPathIsNotUnique = errors.New(`task path is not unique`)

	ErrGitRepoUrlIsNotValid           = errors.New(`git repo url is not valid`)
	ErrGitRepoUrlSchemaIsNotSupported = errors.New(`git repo url schema is not supported`)
)

type Validatable interface {
	Validate() error
}

type Config struct {
	Validatable
	Tasks []*TaskConfig `yaml:"tasks" json:"tasks"`
}

func (c *Config) Validate() error {

	names := make(map[string]bool, len(c.Tasks))
	paths := make(map[string]bool, len(c.Tasks))

	for _, taskConfig := range c.Tasks {
		err := taskConfig.Validate()
		if err != nil {
			return err
		}
		if names[taskConfig.Name] {
			return ErrNameIsNotUnique
		}
		if paths[taskConfig.Path] {
			return ErrPathIsNotUnique
		}
	}
	return nil
}

type TaskConfig struct {
	Validatable
	Name       string `yaml:"name" json:"name"`
	Url        string `yaml:"url" json:"url"`
	Path       string `yaml:"path" json:"path"`
	Insecure   bool   `yaml:"insecure,omitempty" json:"insecure,omitempty"`
	Depth      int    `yaml:"depth,omitempty" json:"depth,omitempty"`
	Submodules bool   `yaml:"submodules,omitempty" json:"submodules,omitempty"`
	Auth       *Auth  `yaml:"auth,omitempty" json:"auth,omitempty"`
	RemoteName string `yaml:"remoteName,omitempty" json:"remoteName,omitempty"`
	Reference  struct {
		Tag    string `yaml:"tag,omitempty" json:"tag,omitempty"`
		Branch string `yaml:"branch,omitempty" json:"branch,omitempty"`
	} `yaml:"reference,omitempty" json:"reference,omitempty"`
	RunOnce         bool  `yaml:"runOnce,omitempty" json:"runOnce,omitempty"`
	IntervalSeconds int   `yaml:"intervalSeconds,omitempty" json:"intervalSeconds,omitempty"`
	Force           bool  `yaml:"force,omitempty" json:"force,omitempty"`
	SingleBranch    *bool `yaml:"singleBranch,omitempty" json:"singleBranch,omitempty"`
	Progress        bool  `yaml:"progress,omitempty" json:"progress,omitempty"`
}

func (c *TaskConfig) Validate() error {
	if len(c.Name) == 0 {
		return ErrNameIsMissing
	}
	if len(c.Path) == 0 {
		return ErrPathIsMissing
	}
	parsedUrl, err := url.Parse(c.Url)
	if err != nil {
		return ErrGitRepoUrlIsNotValid
	}
	urlScheme := parsedUrl.Scheme
	if urlScheme != `http` && urlScheme != `https` {
		return ErrGitRepoUrlSchemaIsNotSupported
	}
	if len(c.Path) == 0 {
		return fmt.Errorf(`target directory (path) is not set`)
	}
	if c.IntervalSeconds < 20 {
		c.IntervalSeconds = 20
	}
	if (len(c.Reference.Branch) > 0) && (len(c.Reference.Tag) > 0) {
		return fmt.Errorf(`you cannot configure branch and tag simultaneously`)
	}
	if c.Auth != nil {
		if err = c.Auth.Validate(); err != nil {
			return err
		}
	}
	return nil
}

type Auth struct {
	Validatable
	BearerToken *Secret `yaml:"bearerToken,omitempty" json:"bearerToken,omitempty"`
	BasicToken  *Secret `yaml:"basicToken,omitempty" json:"basicToken,omitempty"`
	Basic       *Basic  `yaml:"basic,omitempty" json:"basic,omitempty"`
}

func (auth *Auth) Validate() error {
	count := 0
	if auth.BearerToken != nil {
		count++
	}
	if auth.BasicToken != nil {
		count++
	}
	if auth.Basic != nil {
		count++
	}
	if count > 1 {
		return errors.New(`auth -> to many configurations`)
	}
	if auth.BearerToken != nil {
		if err := auth.BearerToken.Validate(); err != nil {
			return fmt.Errorf(`auth -> BearerToken -> %s`, err.Error())
		}
	}
	if auth.BasicToken != nil {
		if err := auth.BasicToken.Validate(); err != nil {
			return fmt.Errorf(`auth -> BasicToken -> %s`, err.Error())
		}
	}
	if auth.Basic != nil {
		if err := auth.Basic.Validate(); err != nil {
			return fmt.Errorf(`auth -> Basic -> %s`, err.Error())
		}
	}
	return nil
}

type Basic struct {
	User     *Secret `yaml:"user,omitempty" json:"user,omitempty"`
	Password *Secret `yaml:"password,omitempty" json:"password,omitempty"`
}

func (basic *Basic) Validate() error {
	if basic.User == nil {
		return errors.New(`user -> not set`)
	}
	if basic.Password == nil {
		return errors.New(`password -> not set`)
	}
	if err := basic.User.Validate(); err != nil {
		return fmt.Errorf(`user -> %s`, err.Error())
	}
	if err := basic.Password.Validate(); err != nil {
		return fmt.Errorf(`password -> %s`, err.Error())
	}
	return nil
}

func (auth *Auth) GitOpts() ([]string, error) {
	if auth.BearerToken != nil {
		token, err := auth.BearerToken.GetValue()
		if err != nil {
			return nil, err
		}

		return []string{
			`-c`,
			fmt.Sprintf(`http.extraHeader=Authorization: Bearer %s`, token),
		}, nil
	}

	if auth.BasicToken != nil {
		token, err := auth.BasicToken.GetValue()
		if err != nil {
			return nil, err
		}
		token = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf(`:%s`, token)))
		return []string{
			`-c`,
			fmt.Sprintf(`http.extraHeader=Authorization: Basic %s`, token),
		}, nil
	}

	if auth.Basic != nil {
		userSecret := auth.Basic.User
		if userSecret == nil {
			return nil, ErrBasicAuthUserIsMissing
		}

		user, err := userSecret.GetValue()
		if err != nil {
			return nil, err
		}

		passwordSecret := auth.Basic.Password
		if passwordSecret == nil {
			return nil, ErrBasicAuthPasswordIsMissing
		}

		password, err := passwordSecret.GetValue()
		if err != nil {
			return nil, err
		}

		token := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf(`%s:%s`, user, password)))
		return []string{
			`-c`,
			fmt.Sprintf(`http.extraHeader=Authorization: Basic %s`, token),
		}, nil
	}

	return []string{}, nil
}

func (auth *Auth) GitAuth() (gitTransport.AuthMethod, error) {

	if auth.BearerToken != nil {
		token, err := auth.BearerToken.GetValue()
		if err != nil {
			return nil, err
		}
		return &gitHttp.TokenAuth{Token: token}, nil
	}

	if auth.BasicToken != nil {
		token, err := auth.BasicToken.GetValue()
		if err != nil {
			return nil, err
		}
		return NewTokenBasicAuth(token), nil
	}

	if auth.Basic != nil {
		userSecret := auth.Basic.User
		if userSecret == nil {
			return nil, ErrBasicAuthUserIsMissing
		}

		user, err := userSecret.GetValue()
		if err != nil {
			return nil, err
		}

		passwordSecret := auth.Basic.Password
		if passwordSecret == nil {
			return nil, ErrBasicAuthPasswordIsMissing
		}

		password, err := passwordSecret.GetValue()
		if err != nil {
			return nil, err
		}

		return &gitHttp.BasicAuth{
			Username: user,
			Password: password,
		}, nil
	}

	return nil, fmt.Errorf(`auth object is not configured`)
}

type Secret struct {
	Value     string           `yaml:"value,omitempty" json:"value,omitempty"`
	ValueFrom *SecretValueFrom `yaml:"valueFrom" json:"valueFrom"`
}

func (secret *Secret) Validate() error {
	count := 0
	if len(secret.Value) > 0 {
		count++
	}
	if secret.ValueFrom != nil {
		count++
	}
	if count != 1 {
		return errors.New(`secret -> value end valueFrom configs cannot be set simultaneously`)
	}
	if secret.ValueFrom != nil {
		if err := secret.ValueFrom.Validate(); err != nil {
			return fmt.Errorf(`valueFrom -> %s`, err.Error())
		}
	}
	return nil
}

type SecretValueFrom struct {
	Env  string `yaml:"env,omitempty" json:"env,omitempty"`
	File string `yaml:"file,omitempty" json:"file,omitempty"`
}

func (valueFrom *SecretValueFrom) Validate() error {
	count := 0
	env := valueFrom.Env
	if len(env) > 0 {
		count++
	}
	file := valueFrom.File
	if len(file) > 0 {
		count++
	}
	if count != 1 {
		return errors.New(`env end file configs cannot be set simultaneously`)
	}
	if len(env) > 0 {
		_, found := os.LookupEnv(env)
		if !found {
			return fmt.Errorf(`env variable %s is not set`, env)
		}
	}
	if len(file) > 0 {
		stat, err := os.Stat(file)
		if err != nil {
			return fmt.Errorf(`file %s is not found`, file)
		}
		if stat.IsDir() {
			return fmt.Errorf(`file %s is directory`, file)
		}
	}
	return nil
}

func (secret *Secret) GetValue() (string, error) {
	var (
		err   error
		bytes []byte
	)

	if len(secret.ValueFrom.File) > 0 {
		bytes, err = os.ReadFile(secret.ValueFrom.File)
		if err == nil {
			return string(bytes), nil
		}
	}

	if len(secret.ValueFrom.Env) > 0 {
		value, found := os.LookupEnv(secret.ValueFrom.Env)
		if found {
			return value, nil
		}
	}

	if len(secret.Value) > 0 {
		return secret.Value, nil
	}

	if err == nil {
		err = fmt.Errorf("secret bad configuration")
	}

	return "", err
}

func (c *TaskConfig) GitCloneCmd() (*exec.Cmd, error) {
	opts := []string{`clone`}
	if c.Insecure {
		opts = append(opts, `-c`, `http.sslVerify=false`)
	}
	if c.Depth > 0 {
		opts = append(opts, `--depth`, fmt.Sprintf(`%d`, c.Depth))
	}
	if c.Submodules {
		opts = append(opts, `--recurse-submodules`)
	}
	if c.Progress {
		opts = append(opts, `--progress`)
	}
	if c.SingleBranch != nil && *c.SingleBranch {
		opts = append(opts, `--single-branch`)
	}
	if len(c.Reference.Branch) > 0 {
		opts = append(opts, `--branch`, c.Reference.Branch)
	} else if len(c.Reference.Tag) > 0 {
		opts = append(opts, `--branch`, c.Reference.Tag)
	}

	if len(c.RemoteName) > 0 {
		log.WithFields(log.Fields{
			`name`:        c.Name,
			`url`:         c.Url,
			`path`:        c.Path,
			`remote_name`: c.RemoteName,
		}).Warn(`manual clone. custom remote name will be ignored`)
	}
	gitOpts, err := c.Auth.GitOpts()
	if err != nil {
		return nil, err
	}
	opts = append(opts, gitOpts...)
	opts = append(opts, c.Url, c.Path)

	cmd := exec.Command(`git`, opts...)
	logWriter := NewLogrusWriter(log.DebugLevel).WithFields(log.Fields{
		`name`: c.Name,
		`url`:  c.Url,
		`path`: c.Path,
	})
	cmd.Stderr = logWriter
	cmd.Stdout = logWriter
	return cmd, nil
}

func (c *TaskConfig) Interval() time.Duration {
	seconds := c.IntervalSeconds
	if seconds <= 0 {
		seconds = defaultIntervalSeconds
	}
	return time.Duration(seconds) * time.Second
}

func (c *TaskConfig) CloneOptions() (*git.CloneOptions, error) {
	var err error
	op := git.CloneOptions{
		URL:             c.Url,
		InsecureSkipTLS: c.Insecure,
	}
	remoteName := c.RemoteName
	if len(remoteName) == 0 {
		remoteName = git.DefaultRemoteName
	}
	op.RemoteName = remoteName
	if len(c.Reference.Tag) > 0 {
		op.ReferenceName = plumbing.NewTagReferenceName(c.Reference.Tag)
	} else if len(c.Reference.Branch) > 0 {
		op.ReferenceName = plumbing.NewBranchReferenceName(c.Reference.Branch)
	}
	if c.Auth != nil {
		if op.Auth, err = c.Auth.GitAuth(); err != nil {
			return nil, err
		}
	}
	if c.Depth > 0 {
		op.Depth = c.Depth
	}
	if c.SingleBranch != nil {
		op.SingleBranch = *c.SingleBranch
	}
	if c.Submodules {
		op.RecurseSubmodules = git.DefaultSubmoduleRecursionDepth
	}

	op.Progress = NewLogrusWriter(log.DebugLevel).WithFields(log.Fields{
		`url`:      c.Url,
		`path`:     c.Path,
		`revision`: op.ReferenceName,
	})

	return &op, nil
}

func isValidReference(reference string) bool {

	return len(reference) > 0 && !strings.HasSuffix(strings.ToLower(reference), `head`)
}

func (c *TaskConfig) PullOptions() (*git.PullOptions, error) {

	_, err := url.Parse(c.Url)
	if err != nil {
		return nil, fmt.Errorf(`GitSyncTask: URL is not valid. %v`, err)
	}

	op := git.PullOptions{
		InsecureSkipTLS: c.Insecure,
	}
	remoteName := c.RemoteName
	if len(remoteName) == 0 {
		remoteName = git.DefaultRemoteName
	}
	op.RemoteName = remoteName
	if isValidReference(c.Reference.Tag) {
		op.ReferenceName = plumbing.NewTagReferenceName(c.Reference.Tag)
	} else if isValidReference(c.Reference.Branch) {
		op.ReferenceName = plumbing.NewBranchReferenceName(c.Reference.Branch)
	}
	if c.Auth != nil {
		if op.Auth, err = c.Auth.GitAuth(); err != nil {
			return nil, err
		}
	}
	if c.Depth > 0 {
		op.Depth = c.Depth
	}
	if c.SingleBranch != nil {
		op.SingleBranch = *c.SingleBranch
	}
	if c.Submodules {
		op.RecurseSubmodules = git.DefaultSubmoduleRecursionDepth
	}

	op.Progress = NewLogrusWriter(log.DebugLevel).WithFields(log.Fields{
		`url`:      c.Url,
		`path`:     c.Path,
		`revision`: op.ReferenceName,
	})

	return &op, nil
}
