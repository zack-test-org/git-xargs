# Release notes drafter

This provides a Golang based AWS Lambda handler that will react to pull request merge hook events from Github to
maintain a release notes draft.

The draft is based on the Gruntwork release notes style guide. Specifically, it maintains the following structure:

```
## Modules Affected

<!-- Contains the list of modules that have been affected by this change -->

## Description

<!-- Contains a description of the change. The drafter will add an entry for each pull request that is a placeholder
  -- TODO, containing the PR title.
  -->

## References

<!-- Contains a list of links to the PRs -->
```


## How it works

This depends on the webhook event from github when a pull request merges. For each PR merge, the handler will:

- Scan the PR diff for the list of modules that have been touched in the changeset.
- Extract a link to the PR.
- Scan for a release in draft stage on the repo.
- If there is a draft, read in the current state as a struct. If not, create a new object.
- Update the draft with the information.
- Upsert the draft to github.

Note that this handler requires synchronization to avoid concurrency issues in updating the draft. We use dynamodb for
this purpose.


## Env Requirements

The release notes draft requires:

- `GITHUB_WEBHOOK_SECRET`: Github Webhook secret
- `GITHUB_API_KEY`: Github API key with scope <TODO>
- AWS IAM profile with DynamoDB access to a lock table. Used for synchronization.
