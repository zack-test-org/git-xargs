# Find GitHub Admin Users

This is a hacky script that:

1. Finds all users with admin permissions in `gruntwork-io` GitHub repos.
1. Prints out any "unexpected" admins. The only users who should have admin access are `brikis98` and `josh-padnick`!

This is useful to check that we are enforcing tight permissions on our repos. In particular, it also ensures that 
virtually no one can bypass or disable [protected branches 
permissions](https://help.github.com/en/github/administering-a-repository/about-protected-branches) on our repos.




## Running the script

First, set your GitHub personal access token as the environment variable `GITHUB_OAUTH_TOKEN`:

```bash
export GITHUB_OAUTH_TOKEN=xxx
```

Then, run the script:

```bash
python find-github-admin-users.py
```

Go get a cup of coffee, as this script takes ~25 min to run (see [how it works](#how-it-works)). At the end,
it'll print out any repos that have unexpected admin users and who those users are.




## How it works

There doesn't seem to be any GitHub API to find all users who have admin access to a repo, so instead, I have to make
a large number of (fairly slow) API calls as follows:

1. [Fetch all repos in the `gruntwork-io` org](https://developer.github.com/v3/repos/#list-organization-repositories).
1. For each repo, [fetch the list of 
   collaborators](https://developer.github.com/v3/repos/collaborators/#list-collaborators). Since we grant all of our
   customers access to our repos, there are thousands of collaborators, but GitHub returns at most 100 per page, so 
   we have to make dozens of API calls per repo to go through all the pages.  