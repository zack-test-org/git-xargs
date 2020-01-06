# `add_remove_github_user_to_team.py` is a script that can be executed as a [Zapier code 
# step](https://zapier.com/help/create/code-webhooks/use-python-code-in-zaps) to add or remove a user to or from a 
# Gruntwork customer's GitHub team. To be able to version and code review this script, it lives in the 
# [Gruntwork prototypes repo](https://github.com/gruntwork-io/prototypes). Every time you want to update the Zap, 
# update the script in this repo first, submit a PR, and when approved, manually copy/paste the updated code into 
# Zapier.

import requests
import logging
import os
import re


def read_from_env(key, required=True):
    """
    Read the given key from the environment. This method first checks the input_data global, which is provided by the
    Zapier code step (https://zapier.com/help/create/code-webhooks/use-python-code-in-zaps). If it's not in input_data,
    the method then looks for an environment variable. If that isn't set either, and required is set to True, this 
    function raises an exception.
    :param key: The key to lookup
    :param required: If set to True and no value is found for the key, raise an Exception 
    :return: The value for the key
    """
    value = read_input_data(key)
    if value:
        return value

    value = os.environ.get(key)
    if value:
        return value

    if required:
        raise Exception('Did not find value for key {} in either input_data or environment variables.'.format(key))
    else:
        return None


def read_input_data(key):
    """
    Reads the given key from the Zapier code step input_data. Note that input_data is magically added by the Zapier code
    step, so this is the only way I could find to look it up that works both when running locally (with no input_data)
    and in Zapier. https://zapier.com/help/create/code-webhooks/use-python-code-in-zaps
    :param key: The key to read from input_data
    :return: The value of the key or None if it's not set
    """
    try:
        return input_data.get(key)
    except:
        return None


def find_github_team(name, github_creds):
    """
    Find a GitHub team with the given name. https://developer.github.com/v3/teams/#get-team    
    :param name: The GitHub team name to lookup 
    :param github_creds: The GitHub creds to use for the API call. Should be a tuple of (username, password). 
    :return: The team ID or None
    """
    logging.info('Looking for a GitHub team called {}'.format(name))
    response = requests.get('https://api.github.com/orgs/gruntwork-io/teams/{}'.format(name), auth=github_creds)
    if response.status_code == 200:
        team_id = response.json()['id']
        logging.info('Found GitHub team with ID {}'.format(team_id))
        return team_id
    else:
        logging.info('No team with name {} found (got response {} from GitHub)'.format(name, response.status_code))
        return None


def get_user_team_membership(github_id, team_id, github_creds):
    """
    Fetch information about the membership of the user with the given GitHub ID on the GitHub team with the given ID.
    https://developer.github.com/v3/teams/members/#get-team-membership    
    :param github_id: The user's GitHub ID
    :param team_id: The GitHub team's ID (NOT the team name)
    :param github_creds: The GitHub creds to use for the API call. Should be a tuple of (username, password). 
    :return: The JSON response from the GitHub get-team-membership API.
    """
    logging.info('Looking up membership info of GitHub user {} on team {}'.format(github_id, team_id))

    url = 'https://api.github.com/teams/{}/memberships/{}'.format(team_id, github_id)
    response = requests.get(url, auth=github_creds)

    if response.status_code == 200:
        logging.info('User {} is a member of team {}'.format(github_id, team_id))
        return response.json()
    elif response.status_code == 404:
        logging.info('User {} is not a member of team {}'.format(github_id, team_id))
        return None        
    else:
        raise Exception('Failed to look up membership for user {} in team {}. Got response {} from GitHub with body: {}.'.format(github_id, team_id, response.status_code, response.json()))


def add_user_to_team(github_id, team_id, github_creds):
    """
    Add the user with the given GitHub ID to the GitHub team with the given ID.
    https://developer.github.com/v3/teams/members/#add-or-update-team-membership    
    :param github_id: The user's GitHub ID 
    :param team_id: The GitHub team's ID (NOT the team name)
    :param github_creds: The GitHub creds to use for the API call. Should be a tuple of (username, password).
    :return: The JSON response from the GitHub add-or-update-team-membership API
    """
    logging.info('Adding GitHub user {} to team {}'.format(github_id, team_id))

    url = 'https://api.github.com/teams/{}/memberships/{}'.format(team_id, github_id)

    payload = {
        'role': 'member'
    }

    response = requests.put(url, auth=github_creds, json=payload)

    if response.status_code == 200:
        logging.info('Successfully added user {} to team {}'.format(github_id, team_id))
        return response.json()
    else:
        raise Exception('Failed to add user {} to team {}. Got response {} from GitHub with body: {}.'.format(github_id, team_id, response.status_code, response.json()))


def remove_user_from_team(github_id, team_id, github_creds):
    """
    Remove the user with the given GitHub ID from the GitHub team with the given ID.
    https://developer.github.com/v3/teams/members/#remove-team-membership    
    :param github_id: The user's GitHub ID
    :param team_id: The GitHub team's ID (NOT the team name)
    :param github_creds: The GitHub creds to use for the API call. Should be a tuple of (username, password).
    :return: An empty object
    """
    logging.info('Deleting GitHub user {} from team {}'.format(github_id, team_id))

    url = 'https://api.github.com/teams/{}/memberships/{}'.format(team_id, github_id)
    response = requests.delete(url, auth=github_creds)

    if response.status_code == 204:
        logging.info('Successfully deleted user {} from team {}'.format(github_id, team_id))
        return {}
    else:
        raise Exception('Failed to delete user {} from team {}. Got response {} from GitHub with body: {}.'.format(github_id, team_id, response.status_code, response.json()))


def format_github_team_name(name):
    """
    Convert the given name to a GitHub team name for a customer. We do this by converting the name to a lower case,
    dash-separated string with a "client-" prefix. E.g., "Foo Bar" becomes "client-foo-bar".
    :param name: The name to dasherize
    :return: The GitHub-friendly team version of name.
    """
    return 'client-{}'.format(re.sub(r'\s', '-', name).lower())


def do_add_user_to_team(github_id, company_name, github_creds):
    """
    Add the user with the given GitHub ID to the GitHub team for the company with the given name.
    :param github_id: The user's GitHub ID
    :param company_name: The name of the company. Will be dasherized and looked up in the GitHub API.
    :param github_creds: The GitHub creds to use for the API call. Should be a tuple of (username, password).
    :return: If the user is already a member, an empty object. Otherwise, the return value of add_user_to_team.
    """
    team_name = format_github_team_name(company_name)
    team_id = find_github_team(team_name, github_creds)

    if not team_id:
        raise Exception('Did not find GitHub team called {}'.format(team_name))

    membership = get_user_team_membership(github_id, team_id, github_creds)
    if membership:
        logging.info('User {} is already a member of team {}. Will not add again.'.format(github_id, team_name))
        return {}

    return add_user_to_team(github_id, team_id, github_creds)


def do_remove_user_from_team(github_id, company_name, github_creds):
    """
    Remove the user with the given GitHub ID from the GitHub team for the company with the given name.
    :param github_id: The user's GitHub ID
    :param company_name: The name of the company. Will be dasherized and looked up in the GitHub API.
    :param github_creds: The GitHub creds to use for the API call. Should be a tuple of (username, password).
    :return: The return value of remove_user_from_team.
    """
    team_name = format_github_team_name(company_name)
    team_id = find_github_team(team_name, github_creds)

    if not team_id:
        raise Exception('Did not find GitHub team called {}'.format(team_name))

    membership = get_user_team_membership(github_id, team_id, github_creds)
    if not membership:
        raise Exception('User {} is not a member of team {}'.format(github_id, team_name))

    return remove_user_from_team(github_id, team_id, github_creds)


def is_affirmative_value(value):
    """
    Returns true if the given value represents an "affirmative" from the user (i.e., a "Yes")
    :param value:
    :return:
    """
    return value.lower() in ["yes", "true", "1"]


def is_negative_value(value):
    """
    Returns true if the given value represents an "negative" from the user (i.e., a "No")
    :param value:
    :return:
    """
    return value.lower() in ["no", "false", "0"]


def run():
    """
    Main entrypoint for the code. Reads data from the environment and adds or removes the specified user to/from the
    specified team. Returns the response body of the GitHub add user or delete user GitHub API call.
    """
    github_user = read_from_env('GITHUB_USER')
    github_pass = read_from_env('GITHUB_TOKEN')

    assert len(github_user) > 2, 'GitHub username does not seem to be valid (less than 3 characters long)'
    assert len(github_pass) > 2, 'GitHub password does not seem to be valid (less than 3 characters long)'

    github_creds = (github_user, github_pass)

    github_id = read_from_env('github_id')
    user_active = read_from_env('user_active', required=False)

    assert len(github_id) > 2, 'GitHub ID does not seem to be valid (less than 3 characters long)'

    company_name = read_from_env('company_name')
    company_current_users = int(read_from_env('company_current_users'))
    company_max_users = int(read_from_env('company_max_users'))
    company_active = read_from_env('company_active')

    assert len(company_name) > 2, 'First name does not seem to be valid (less than 3 characters long)'
    assert company_current_users >= 0, 'Company current users should not be negative'
    assert company_max_users >= 0, 'Company max users should not be negative'
    assert company_current_users <= company_max_users, 'Company current user must not exceed company max users'

    if not is_affirmative_value(company_active):
        raise Exception('Company {} is not active, cannot add or remove users!'.format(company_name))

    if is_affirmative_value(user_active):
        logging.info('The "active" input for the user is set to "Yes", so adding user to team.')
        return do_add_user_to_team(github_id, company_name, github_creds)
    elif is_negative_value(user_active):
        logging.info('The "active" input for the user is set to "No", so removing user from the team.')
        return do_remove_user_from_team(github_id, company_name, github_creds)
    else:
        logging.info('The "active" input for the user is not set to "Yes" or "No", so assuming this entry is still a WIP and will not take any action.')
        return {}


# Zapier requires that you set a variable called output with your returned data
logging.basicConfig(format='%(asctime)s [%(levelname)s] %(message)s', level=logging.INFO)
output = run()
logging.info(output)
