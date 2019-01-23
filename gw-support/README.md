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

The Oauth flow is managed with a http server that will be run in the background. This server will cache the credentials
in memory.
