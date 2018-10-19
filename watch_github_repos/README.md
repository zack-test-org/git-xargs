# Watch Github Repos

A script to watch all the repos in an organization. 

This is a refined version of https://gist.github.com/thet/c1ce413bdabc771cba1b


## Usage

- Get a Github personal access token with `repo` and `org` scope.
- Set the environment variable `GITHUB_OAUTH_TOKEN`
- Install dependencies: `pip install -r requirements.txt`
- Create log file: `touch repos_set.txt`
- Run: `python watch_github.py gruntwork-io`
	- NOTE: Use `--dry` to see which repos this will subscribe you to
