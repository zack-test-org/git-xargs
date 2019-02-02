# gw-support

A tool to help find out the state of the support rotation at Gruntwork. Specifically you can use the tool to find out:

- who is currently on support
- the next N weeks of the support rotation
- when you will be on support next


## How does it work

This tool assumes that there is a calendar that contains the support rotation schedule. This tool also assumes that the
rotation event is setup with the prefix `Support:`.

Given that, the tool obtains Oauth credentials with read only scope to query the Gsuites calendar API to retrieve
visible events. Using the credentials, the tool will scan for the support events on the calendar and answer the queries
based on the information it finds.

### The local server

To manage the Oauth flow, which requires browser interaction, we spawn a local HTTP server on port 56789 in the
background. This server is used to initiate the Oauth flow by providing the Google login URL, which the CLI will use to
open a browser tab.

In that new tab, you will be asked to login to Google and authorize the CLI read only access to the calendar events.
Once you authorize, the oauth flow will complete, posting the authorized token to the local server. This token is then
used by the CLI to access the Google calendar events.

By doing so, the CLI is able to run use your token to check the events on your calendar to find when you are on support
next.


### Local server CSRF

To prevent retrieving the token from a random browser page, the local http server protects the credentials page using
basic auth managed by a file stored on the local hard disk. This token is generated when the server starts up and is
accessible by the client, but is not accessible to the browser, thus preventing unauthorized access to the credentials
from a random web page access.


## Usage

- See who is on support now
    ```
    gw-support now
    ```

- See when you are on support next
    ```
    gw-support next
    ```
