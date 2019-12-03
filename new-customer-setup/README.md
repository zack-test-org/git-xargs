# New Customer Setup

`create_github_team.py` is a script that can be executed as a [Zapier code 
step](https://zapier.com/help/create/code-webhooks/use-python-code-in-zaps) to create a team in GitHub for a Gruntwork
customer and to grant that team access to all the repos it should have access to. To be able to version and code 
review this script, it lives in the [Gruntwork prototypes repo](https://github.com/gruntwork-io/prototypes). Every 
time you want to update the Zap, update the script in this repo first, submit a PR, and when approved, manually 
copy/paste the updated code into Zapier.




## Run locally

This script requires Python 3+.

To run locally, first install dependencies:

```bash
pip3 install -r requirements.txt --user
```

Then, configure environment variables for the values that Zapier would've passed in as inputs:

```bash
export GITHUB_USER=xxx       # GitHub user name to use for auth
export GITHUB_TOKEN=xxx      # GitHub personal access token to use for auth
export company_name=xxx      # The name of the company for which to create a GitHub team
export subscription_type=xxx # The company's subscription type. Must be one of: aws, gcp, enterprise.
export active=yes            # Set to "yes" to indicate the company is active and a GitHub team should be created.
```

Finally, run the script:

```bash
python3 create_github_team.py
```




## Run in Zapier

To run in Zapier, copy/paste the contents of `create_github_team.py` directly into the Zap and click the 
"Test & Review" button in the UI to run it.