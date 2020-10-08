# Test PR

Our repos do not run external PRs automatically for security reasons
i.e. leaking of environment variables from circle environment.
This means often tests are run after merging the PR which is not an ideal
situation. This script contains a series of steps to push the PR changes to
another branch in the repo, kicking off CI tests.

This script deletes the local branch and the remote once it's pushed, but
after CI runs you still have to manually remove the branch on origin. This
is a simple initial version and therefore we are not waiting on our CI to do
that automatically here yet.

Usage:

```
export GITHUB_OAUTH_TOKEN=<YOUR PERSONAL ACCESS TOKEN>
cd local_repo_clone
testpr --pr 42
```

Dependencies:
- You need `jq` installed locally (e.g. `brew install jq` on macOS)
- Make sure the git remote that points to the gruntwork-io GH repo is named 'origin'

