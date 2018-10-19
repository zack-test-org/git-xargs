# Watch Github Repos

A script to watch all the repos in an organization. This way, you get notifications for all PRs & issues in that repo,
so you don't miss any activity. At Gruntwork, it is important that we keep track of issues, comments, and pull requests
from our customers, as that is a significant source of feedback from our customer base.

This is a refined version of https://gist.github.com/thet/c1ce413bdabc771cba1b


## Usage

- Get a Github personal access token with `repo` and `org` scope.
- Set the environment variable `GITHUB_OAUTH_TOKEN`
- Install dependencies: `pip install -r requirements.txt`
- Run: `python watch_github.py gruntwork-io`
	- NOTE: Use `--dry` to see which repos this will subscribe you to
