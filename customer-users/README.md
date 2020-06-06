# Customer Authorized Users Lookup

This CLI utility can be used to lookup all the active and inactive authorized users for a customer that is managed with
[the Gruntwork Customers
spreadsheet](https://docs.google.com/spreadsheets/d/1vvUoSZxoGhWVQhyFbceRsFTbSi3jt-0MYKDgGBSt6Jc/edit). You can use this
to generate a table that you can provide to customers that ask for an audit of all the current active users of a
subscription.

This works similar to the
[refarch-init](https://github.com/gruntwork-io/usage-patterns/tree/master/scripts/refarch-init) script for the Reference
Architecture:

This tool uses the spreadsheets API to retrieve information from the Gruntwork Customers spreadsheet. Note that this
will obtain Oauth credentials with read only scope from the user (NOTE: the oauth client ID is shared with
`refarch-init`, and the app name in Google is `refarch-deployment-initializer`).

To manage the Oauth flow, which requires browser interaction, we spawn a local HTTP server on port 46548 in the
background. This server is used to initiate the Oauth flow by providing the Google login URL, which the CLI will use to
open a browser tab.

In that new tab, you will be asked to login to Google and authorize the CLI read only access Google Spreadsheets.
Once you authorize, the oauth flow will complete, posting the authorized token to the local server. This token is then
used by the CLI to access the Google Spreadsheet containing Reference Architecture Questionnaire responses.


### Local server CSRF

To prevent retrieving the token from a random browser page, the local http server protects the credentials page using
basic auth managed by a file stored on the local hard disk. This token is generated when the server starts up and is
accessible by the client, but is not accessible to the browser, thus preventing unauthorized access to the credentials
from a random web page access.


## Usage

To use the script, run `go run . lookup`. This will download all dependencies, build the CLI, and run it.


## Troubleshooting

### Issues accessing private dependencies

This module depends on private repos, which requires authentication credentials when fetching the dependencies in `go`.
If go is not authenticated to github, you may get an error like below:

```
go: github.com/gruntwork-io/houston-cli@v0.0.20: reading github.com/gruntwork-io/houston-cli/go.mod at revision v0.0.20: unknown revision v0.0.20
```

To allow go to fetch the dependencies, follow these steps to cache the github credentials in your toolchain:

**Mac OSX**

- Generate a personal access token with repo access.
- Follow the steps in [the Github docs](https://help.github.com/en/github/using-git/caching-your-github-password-in-git)
  to setup the `osxtoolchain` as the git credentials cache.
- Clone any private repo, and enter the personal access token as your password.
- Now go should be able to retrieve the credentials from the keychain to clone the private repos.


**Linux**

TODO
