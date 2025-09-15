# This is how `actions/checkout@v4` works

```
$ /usr/bin/git version
$ /usr/bin/git config --global --add safe.directory /home/runner/work/scan-action-test/scan-action-test
$ /usr/bin/git init /home/runner/work/scan-action-test/scan-action-test
$ /usr/bin/git remote add origin https://github.com/gregory-entro/scan-action-test
$ /usr/bin/git config --local gc.auto 0

# Skipping setting up auth

$ /usr/bin/git -c protocol.version=2 fetch --no-tags --prune --no-recurse-submodules --depth=1 origin +c5f6362ad25a5feaa5b578b42f4f4f79ff17606f:refs/remotes/pull/2/merge
$ /usr/bin/git sparse-checkout disable
$ /usr/bin/git config --local --unset-all extensions.worktreeConfig

/usr/bin/git checkout --progress --force refs/remotes/pull/2/merge

$ /usr/bin/git log -1 --format=%H
```