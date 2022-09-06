# git-sync-go

## Description

TBD

## App Usage

```
NAME:
   git-sync - A new cli application

USAGE:
   git-sync [global options] command [command options] [arguments...]

VERSION:
   local

AUTHOR:
   Dmytro Horkhover <gd.mail.89@gmail.com>

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help (default: false)
   --version, -v  print the version (default: false)
   
   config
   --config value  path to the yaml config file [$CONFIG, $GIT_SYNC_CONFIG]
   
   logs
   --log-colors        Log Colors (for "logfmt" format) (default: false) [$LOG_COLORS, $GIT_SYNC_LOG_COLORS]
   --log-format value  Log Format: json, logfmt (default: logfmt) [$LOG_FORMAT, $GIT_SYNC_LOG_FORMAT]
   --log-level value   Log Level: panic, fatal, error, warn, info, debug, trace (default: info) [$LOG_LEVEL, $GIT_SYNC_LOG_LEVEL]
   --log-pretty        Pretty Log Format (for "json" format) (default: false) [$LOG_PRETTY, $GIT_SYNC_LOG_PRETTY]
   
   metrics
   --port value  HTTP Server Port (default: 9125) [$GIT_SYNC_PORT]
   --server      Start HTTP Webserver (default: false) [$GIT_SYNC_SERVER]

```

## Config File Example

```yaml
tasks:
  - name: leetcode-go
    url: https://github.com/DimkaGorhover/leetcode-go.git
    path: /path/to/dir/leetcode-go
    depth: 1
    auth:
      basic:
        user:
          valueFrom:
            env: GIT_USER
        password:
          valueFrom:
            file: /run/secrets/password

  - name: leetcode-java
    url: https://github.com/DimkaGorhover/leetcode-java.git
    path: /path/to/dir/leetcode-java
    depth: 1
    auth:
      basicToken:
        valueFrom:
          env: PAT
```

## Links

- [Github: go-git](https://github.com/go-git/go-git)
- [Embedding Git in your Applications - go-git](https://git-scm.com/book/en/v2/Appendix-B%3A-Embedding-Git-in-your-Applications-go-git)
