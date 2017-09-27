# Fast bulk git repository mirroring

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

## License

Software licensed under the [MIT License](http://www.opensource.org/licenses/MIT).
