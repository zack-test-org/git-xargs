# `add_github_user_to_team.py` is a script that can be executed as a [Zapier code 
# step](https://zapier.com/help/create/code-webhooks/use-python-code-in-zaps) to add a user to a Gruntwork customer's
# GitHub team. To be able to version and code review this script, it lives in the 
# [Gruntwork prototypes repo](https://github.com/gruntwork-io/prototypes). Every time you want to update the Zap, 
# update the script in this repo first, submit a PR, and when approved, manually copy/paste the updated code into 
# Zapier.

import requests
import logging
import os
import re


# Read the given key from the environment. This method first checks the input_data global, which is provided by the
# Zapier code step (https://zapier.com/help/create/code-webhooks/use-python-code-in-zaps). If it's not in input_data,
# the method then looks for an environment variable. If that isn't set either, this method raises an exception.
def read_from_env(key):
    value = read_input_data(key)
    if value:
        return value

    value = os.environ.get(key)
    if value:
        return value

    raise Exception('Did not find value for key %s in either input_data or environment variables.' % key)


# Reads the given key from the Zapier code step input_data. Note that input_data is magically added by the Zapier code
# step, so this is the only way I could find to look it up that works both when running locally (with no input_data)
# and in Zapier. https://zapier.com/help/create/code-webhooks/use-python-code-in-zaps
def read_input_data(key):
    try:
        return input_data.get(key)
    except:
        return None


# Find a GitHub team with the given name. Returns the team ID or None.
# https://developer.github.com/v3/teams/#get-team
def find_github_team(name, github_creds):
    logging.info('Looking for a GitHub team called %s' % name)
    response = requests.get('https://api.github.com/orgs/gruntwork-io/teams/%s' % name, auth=github_creds)
    if response.status_code == 200:
        team_id = response.json()['id']
        logging.info('Found GitHub team with ID %s' % team_id)
        return team_id
    else:
        logging.info('No team with name %s found (got response %d from GitHub)' % (name, response.status_code))
        return None


# Add the user with the given GitHub ID to the GitHub team with the given ID
# https://developer.github.com/v3/teams/members/#add-or-update-team-membership
def add_user_to_team(github_id, team_id, github_creds):
    logging.info('Adding GitHub user %s to team %s' % (github_id, team_id))

    url = 'https://api.github.com/teams/%s/memberships/%s' % (team_id, github_id)

    payload = {
        'role': 'member'
    }

    response = requests.put(url, auth=github_creds, json=payload)

    if response.status_code == 200:
        logging.info('Successfully added user %s to team %s' % (github_id, team_id))
        return response.json()
    else:
        raise Exception('Failed to add user %s to team %s. Got response %d from GitHub witt body: %s.' % (github_id, team_id, response.status_code, response.json()))


# Convert the given name to a lower case, dash-separated string. E.g., "Foo Bar" becomes "foo-bar".
def dasherize(name):
    return re.sub(r'\s', '-', name).lower()


# Add the user with the given GitHub ID to the GitHub team for the company with the given name
def add_user_to_team_if_necessary(github_id, company_name, github_creds):
    team_name = dasherize(company_name)
    team_id = find_github_team(team_name, github_creds)

    if not team_id:
        raise Exception('Did not find GitHub team called %s' % team_name)

    return add_user_to_team(github_id, team_id, github_creds)


# Main entrypoint for the code. Reads data from the environment and creates the GitHub team. Returns the response body
# of the GitHub create team API call.
def run():
    github_user = read_from_env('GITHUB_USER')
    github_pass = read_from_env('GITHUB_TOKEN')

    assert len(github_user) > 2, 'GitHub username does not seem to be valid (less than 3 characters long)'
    assert len(github_pass) > 2, 'GitHub password does not seem to be valid (less than 3 characters long)'

    github_creds = (github_user, github_pass)

    github_id = read_from_env('github_id')
    user_active = read_from_env('user_active')

    assert len(github_id) > 2, 'GitHub ID does not seem to be valid (less than 3 characters long)'

    company_name = read_from_env('company_name')
    company_current_users = int(read_from_env('company_current_users'))
    company_max_users = int(read_from_env('company_max_users'))
    company_active = read_from_env('company_active')

    assert len(company_name) > 2, 'First name does not seem to be valid (less than 3 characters long)'
    assert company_current_users >= 0, 'Company current users should not be negative'
    assert company_max_users >= 0, 'Company max users should not be negative'
    assert company_current_users <= company_max_users, 'Company current user must not exceed company max users'

    if company_active != 'Yes':
        raise Exception('Company %s is not active, cannot add more users!' % company_name)

    if user_active == "Yes":
        logging.info('The "active" input for the user is set to "Yes", so adding user to team.')
        return add_user_to_team_if_necessary(github_id, company_name, github_creds)
    elif company_active == "No":
        raise Exception('The "active" input for the user is is set to "No", but user deletion has not been implemented yet!')
    else:
        logging.info('The "active" input for the user is not set to "Yes" or "No", so assuming this entry is still a WIP and will not take any action.')
        return {}


# Zapier requires that you set a variable called output with your returned data
logging.basicConfig(format='%(asctime)s [%(levelname)s] %(message)s', level=logging.INFO)
output = run()
logging.info(output)
