# Release notes drafter

This provides a Golang based AWS Lambda handler that will react to pull request merge hook events from Github to
maintain a release notes draft on the repo.

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


## How to deploy

- Build and deploy as a public lambda object with an API gateway (TODO: see if github has static ips we can use for ip
  whitelisting)
- Set the runtime environment variables:
    * `GITHUB_WEBHOOK_SECRET`: Github Webhook secret. This should be generated.
    * `GITHUB_API_KEY`: Github API key with scope `repo`. This should be for a user with enough permissions to update
      the release notes on the repo (Read/Write access).
    * AWS IAM profile with DynamoDB access to a lock table. Used for synchronization.

- Setup a webhook for each repo that we want the release notes drafter to handle. The webhook should point to the API
  gateway endpoint and the secret key should be the one set in the runtime environment for the lambda function.


## How to test locally

- Set `IS_LOCAL` environment variable
- Build and run the app
- This will run a web server on port 8080
- Start [`ngrok`](https://ngrok.com/) to expose the app
- Setup github repo with webhook to point to the ngrok endpoint
