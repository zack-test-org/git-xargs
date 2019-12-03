# New Customer Setup

This repo contains two scripts that are meant to be executed as [Zapier code 
steps](https://zapier.com/help/create/code-webhooks/use-python-code-in-zaps) to help set up Gruntwork customers: 

1. `create_github_team.py`: Creates a new team in the Gruntwork GitHub org and grants that team access to the 
   appropriate Git repos.
1. `add_github_user_to_team.py`: Add a user to a GitHub team.    

These scripts work in conjunction with the [Gruntwork Customers Google 
Sheet](https://docs.google.com/spreadsheets/d/1vvUoSZxoGhWVQhyFbceRsFTbSi3jt-0MYKDgGBSt6Jc/edit#gid=0). Each time a
new company is added to the sheet, Zapier runs `create_github_team.py` to create a GitHub team for that company. Each
time a new user is added to the sheet, Zapier runs `add_github_user_to_team.py` to add the GitHub user to the 
appropriate team. In the future, we'll be able to implement deletion too based on the `active` status of a user or 
company being set to `No` in the Google Sheet, but we have not added that yet.

To be able to version and code review these scripts, they live in this Gruntwork prototypes repo. Every time you want 
to update a Zap, update the corresponding script in this repo first, submit a PR, and when approved, manually 
copy/paste the updated code into Zapier.




## Local setup

Both scripts require Python 3+.

To run them locally, you need to install dependencies:

```bash
pip3 install -r requirements.txt --user
```




## Run `create_github_team.py` locally

Configure environment variables for the values that Zapier would've passed in as inputs:

```bash
export GITHUB_USER=xxx       # GitHub user name to use for auth
export GITHUB_TOKEN=xxx      # GitHub personal access token to use for auth
export company_name=xxx      # The name of the company. The GitHub team will use a dasherized version of this name.
export subscription_type=xxx # The company's subscription type. Must be one of: aws, gcp, enterprise.
export active=yes            # Set to "yes" to indicate the company is active and a GitHub team should be created.
```

Run the script:

```bash
python3 create_github_team.py
```




## Run `add_github_user_to_team.py` locally

Configure environment variables for the values that Zapier would've passed in as inputs:

```bash
export GITHUB_USER=xxx            # GitHub user name to use for auth
export GITHUB_TOKEN=xxx           # GitHub personal access token to use for auth
export github_id=xxx              # The GitHub ID of the user to add
export user_active=xxx            # Set to "yes" to indicate the user is active and should be added to the team.
export company_name=xxx           # The name of the company. The GitHub team will be found by dasherizing this name.
export company_current_users=xxx  # The number of users the company currently has.
export company_max_users=xxx      # The max users the company can have per its contract.
export company_active=yes         # Set to "yes" to indicate the company is active and more users can be added.
```

Run the script:

```bash
python3 add_github_user_to_team.py
```




## Run the scripts in Zapier

To run in Zapier, copy/paste the contents of the scripts directly into the corresponding Zaps and click the 
"Test & Review" button in the UI to run it.