# New Customer Setup

This repo contains two scripts that are meant to be executed as [Zapier code 
steps](https://zapier.com/help/create/code-webhooks/use-python-code-in-zaps) to help set up Gruntwork customers: 

1. `create_remove_github_team.py`: Creates removes a team in the Gruntwork GitHub org. When creating a team, also 
   grants that team access to the appropriate Git repos.
1. `add_remove_github_user_to_team.py`: Add or remove a user to or from a GitHub team.    

These scripts work in conjunction with the [Gruntwork Customers Google 
Sheet](https://docs.google.com/spreadsheets/d/1vvUoSZxoGhWVQhyFbceRsFTbSi3jt-0MYKDgGBSt6Jc/edit#gid=0): 

- Each time a new company is added to the sheet with `active` status set to `Yes`, Zapier runs 
  `create_remove_github_team.py` to create a GitHub team for that company. 
- Each company's `active` status set to `No`, Zapier runs `create_remove_github_team.py` to remove the GitHub team for 
  that company. 
- Each time a new user is added to the sheet with `active` status set to `Yes`, Zapier runs 
  `add_remove_github_user_to_team.py` to add the GitHub user to the appropriate team. 
- Each time a user's `active` status is changed to `Yes`, Zapier runs `add_remove_github_user_to_team.py` to remove the 
  GitHub user from the appropriate team. 

To be able to version and code review these scripts, they live in this Gruntwork prototypes repo. Every time you want 
to update a Zap, update the corresponding script in this repo first, submit a PR, and when approved, manually 
copy/paste the updated code into Zapier.




## Local setup

1. Both scripts require Python 3+.
1. To run the scripts locally, you first need to install dependencies:

    ```bash
    pip3 install -r requirements.txt --user
    ```
1. You will need a [GitHub personal access 
   token](https://help.github.com/en/github/authenticating-to-github/creating-a-personal-access-token-for-the-command-line)
   with `admin:org` permissions. You must also have admin access to all repos in the Gruntwork IaC Library and/or be an
   owner of the `gruntwork-io` GitHub org.




## Run `create_remove_github_team.py` locally

Configure environment variables for the values that Zapier would've passed in as inputs:

```bash
export GITHUB_USER=xxx       # GitHub user name to use for auth
export GITHUB_TOKEN=xxx      # GitHub personal access token to use for auth
export company_name=xxx      # The name of the company. The GitHub team will use a dasherized version of this name.
export subscription_type=xxx # The company's subscription type. Must be one of: aws, gcp, enterprise, aws-cis, enterprise-cis.
export active=Yes            # Set to "Yes" to add a GitHub team for the company and "No" to remove the GitHub team.
```

Run the script:

```bash
python3 create_remove_github_team.py
```




## Run `add_remove_github_user_to_team.py` locally

Configure environment variables for the values that Zapier would've passed in as inputs:

```bash
export GITHUB_USER=xxx            # GitHub user name to use for auth
export GITHUB_TOKEN=xxx           # GitHub personal access token to use for auth
export github_id=xxx              # The GitHub ID of the user to add
export user_active=xxx            # Set to "yes" to indicate the user is active and should be added to the team.
export company_name=xxx           # The name of the company. The GitHub team will be found by dasherizing this name.
export company_current_users=xxx  # The number of users the company currently has.
export company_max_users=xxx      # The max users the company can have per its contract.
export company_active=Yes         # Set to "Yes" to add the user to the GitHub team and "No" to remove them from the team.
```

Run the script:

```bash
python3 add_remove_github_user_to_team.py
```




## Run the scripts in Zapier

To run in Zapier, copy/paste the contents of the scripts directly into the corresponding Zaps and click the
"Test & Review" button in the UI to run it.

## Working with the Gruntwork Customers Google Sheet

This section gives information about working with the [Gruntwork Customers Google Sheet](https://docs.google.com/spreadsheets/d/1vvUoSZxoGhWVQhyFbceRsFTbSi3jt-0MYKDgGBSt6Jc/edit#gid=0):

### Migrating a Customer from `github-terraform`

1. Add the company on the Company tab of the sheet. Enter all the info, but do NOT set “Run Zapier” to “Yes”. Instead, set it to `(migrated-from-github-terraform)` to indicate the company already exists. Zapier will ignore this row, which is correct, as there’s nothing to do.
1. Add each user to the Users tab of the sheet. Enter all the info for the user, but do NOT set “Run Zapier” to “Yes”. Instead, set it to `(migrated-from-github-terraform)` to indicate the user already exists. Zapier will ignore this row, which is correct, as there’s nothing to do. You may have to look up each user’s GitHub ID, name, and email in different places: e.g., GitHub ID in github-terraform, the rest in ZenDesk, Slack, or the New Customer Setup Form (i.e., when the customer requested the user be added), or possibly Teachable.
1. Delete the company’s folder from github-terraform and submit a PR with the changes.
1. Now you can update the company in the Gruntwork Customers Sheet from here on out. Add or remove users, set “Run in Zapier” to “Yes”, and Zapier will do the rest.
