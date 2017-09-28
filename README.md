# Fast bulk git repository mirroring

[![Build Status](https://travis-ci.org/arkady-emelyanov/git-mirror.svg?branch=master)](https://travis-ci.org/arkady-emelyanov/git-mirror)
[![Go Report Card](https://goreportcard.com/badge/github.com/arkady-emelyanov/git-mirror)](https://goreportcard.com/report/github.com/arkady-emelyanov/git-mirror)
[![codebeat badge](https://codebeat.co/badges/379bd888-75ca-4b77-8a1e-0550d1652fd6)](https://codebeat.co/projects/github-com-arkady-emelyanov-git-mirror-master)

Small and fast utility to mirror git repositories.

## Usage

Make sure all repositories are reachable via `git` command-line utility

Create file `repository_list.json` with following content:
```json
[
  {
    "server1": "git@server1.example.com:path/repo1.git",
    "server2": "https://server2.example.com/another_path/repo1.git"
  },
  {
    "server1": "git@server1.example.com:path/repo2.git",
    "server2": "https://server2.example.com/another_path/repo2.git"
  }
]
```

Perform mirroring:
```bash
git-mirror \
    -cacheDir=/path/to/cache -concurrency=10 \
    repository_list.json \
    server1 server2
```

All repositories from `server1` key, will be mirrored to repositories of `server2`.

If `cacheDir` is not provided, repositories will be cloned to a temporary location and 
will be removed up after mirroring.

`concurrency` flag allows to tune number of parallel workers.

## License

Software licensed under the [MIT License](http://www.opensource.org/licenses/MIT).
