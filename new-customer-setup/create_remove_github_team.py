# `create_remove_github_team.py` is a script that can be executed as a [Zapier code
# step](https://zapier.com/help/create/code-webhooks/use-python-code-in-zaps) to create or remove a team in GitHub for
# a Gruntwork customer. To be able to version and code review this script, it lives in the
# [Gruntwork prototypes repo](https://github.com/gruntwork-io/prototypes). Every time you want to update the Zap,
# update the script in this repo first, submit a PR, and when approved, manually copy/paste the updated code into
# Zapier.

import requests
import logging
import os
import re

aws_repos = [
    'gruntwork-io/bash-commons',
    'gruntwork-io/cloud-nuke',
    'gruntwork-io/fetch',
    'gruntwork-io/gruntkms',
    'gruntwork-io/gruntwork',
    'gruntwork-io/gruntwork-cli',
    'gruntwork-io/gruntwork-installer',
    'gruntwork-io/helm-kubernetes-services',
    'gruntwork-io/infrastructure-as-code-training',
    'gruntwork-io/infrastructure-live-acme',
    'gruntwork-io/infrastructure-live-multi-account-acme',
    'gruntwork-io/infrastructure-modules-acme',
    'gruntwork-io/infrastructure-modules-multi-account-acme',
    'gruntwork-io/intro-to-terraform',
    'gruntwork-io/kubergrunt',
    'gruntwork-io/module-asg',
    'gruntwork-io/module-aws-monitoring',
    'gruntwork-io/module-cache',
    'gruntwork-io/module-ci',
    'gruntwork-io/module-data-storage',
    'gruntwork-io/module-ecs',
    'gruntwork-io/module-load-balancer',
    'gruntwork-io/module-security',
    'gruntwork-io/module-server',
    'gruntwork-io/module-vpc',
    'gruntwork-io/package-beanstalk',
    'gruntwork-io/package-elk',
    'gruntwork-io/package-kafka',
    'gruntwork-io/package-lambda',
    'gruntwork-io/package-messaging',
    'gruntwork-io/package-openvpn',
    'gruntwork-io/package-sam',
    'gruntwork-io/package-static-assets',
    'gruntwork-io/package-terraform-utilities',
    'gruntwork-io/package-zookeeper',
    'gruntwork-io/package-mongodb',
    'gruntwork-io/sample-app-backend-acme',
    'gruntwork-io/sample-app-backend-multi-account-acme',
    'gruntwork-io/sample-app-backend-packer',
    'gruntwork-io/sample-app-frontend-acme',
    'gruntwork-io/sample-app-frontend-multi-account-acme',
    'gruntwork-io/terraform-aws-couchbase',
    'gruntwork-io/terraform-aws-eks',
    'gruntwork-io/terraform-kubernetes-helm',
    'gruntwork-io/terratest',
    'gruntwork-io/terragrunt',
    'gruntwork-io/terragrunt-infrastructure-live-example',
    'gruntwork-io/terragrunt-infrastructure-modules-example',
    'gruntwork-io/toc',
    'hashicorp/terraform-aws-consul',
    'hashicorp/terraform-aws-nomad',
    'hashicorp/terraform-aws-vault'
]

gcp_repos = [
    'gruntwork-io/bash-commons',
    'gruntwork-io/cloud-nuke',
    'gruntwork-io/fetch',
    'gruntwork-io/gruntwork-cli',
    'gruntwork-io/gruntwork-installer',
    'gruntwork-io/helm-kubernetes-services',
    'gruntwork-io/infrastructure-as-code-training',
    'gruntwork-io/infrastructure-live-google',
    'gruntwork-io/infrastructure-modules-google',
    'gruntwork-io/intro-to-terraform',
    'gruntwork-io/kubergrunt',
    'gruntwork-io/module-ci',
    'gruntwork-io/module-security',
    'gruntwork-io/terraform-kubernetes-helm',
    'gruntwork-io/terratest',
    'gruntwork-io/terragrunt',
    'gruntwork-io/terragrunt-infrastructure-live-example',
    'gruntwork-io/terragrunt-infrastructure-modules-example',
    'gruntwork-io/toc',
    'hashicorp/terraform-google-consul',
    'hashicorp/terraform-google-nomad',
    'hashicorp/terraform-google-vault'
]

repos_for_subscription = {
    'aws': aws_repos,
    'gcp': gcp_repos,
    'enterprise': list(set(aws_repos + gcp_repos))
}


def read_from_env(key, required=True):
    """
    Read the given key from the environment. This method first checks the input_data global, which is provided by the
    Zapier code step (https://zapier.com/help/create/code-webhooks/use-python-code-in-zaps). If it's not in input_data,
    the method then looks for an environment variable. If that isn't set either, and required is set to True, this
    method raises an exception.
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
        raise Exception('Did not find value for key %s in either input_data or environment variables.' % key)
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
    Find a GitHub team with the given name. Returns the team ID or None.
    https://developer.github.com/v3/teams/#get-team
    :param name: The name of the GitHub team
    :param github_creds: The GitHub creds to use for the API call. Should be a tuple of (username, password).
    :return: The ID of the team or None
    """
    logging.info('Looking for a GitHub team called %s' % name)
    response = requests.get('https://api.github.com/orgs/gruntwork-io/teams/%s' % name, auth=github_creds)
    if response.status_code == 200:
        team_id = response.json()['id']
        logging.info('Found GitHub team with ID %s' % team_id)
        return team_id
    else:
        logging.info('No team with name %s found (got response %d from GitHub)' % (name, response.status_code))
        return None


def create_github_team(name, description, repos, github_creds):
    """
    Create a GitHub team with the given name and description. https://developer.github.com/v3/teams/#create-team
    :param name: The name of the GitHub team to create
    :param description: A description of the team
    :param repos: The repos the team should get access to
    :param github_creds: The GitHub creds to use for the API call. Should be a tuple of (username, password).
    :return: Returns the JSON body of the GitHub create-team API response
    """
    logging.info('Creating new GitHub team called %s and granting it access to repos: %s' % (name, repos))

    payload = {
        'name': name,
        'description': description,
        'privacy': 'secret',
        'repo_names': repos
    }

    response = requests.post('https://api.github.com/orgs/gruntwork-io/teams', auth=github_creds, json=payload)

    if response.status_code == 201:
        response_body = response.json()
        team_id = response_body['id']
        logging.info('Successfully created GitHub team called %s with ID %s' % (name, team_id))
        return response_body
    else:
        raise Exception('Failed to create team called %s. Got response %d from GitHub with body: %s.' % (name, response.status_code, response.json()))


def remove_github_team(name, team_id, github_creds):
    """
    Remove a GitHub team with the given name and description.
    https://developer.github.com/v3/teams/#delete-team
    :param name: The name of the team to remove. Solely used for logging.
    :param team_id: The ID of the team to remove (NOT the name)
    :param github_creds: The GitHub creds to use for the API call. Should be a tuple of (username, password).
    :return: An empty object
    """
    logging.info('Deleting GitHub team called %s with ID %s' % (name, team_id))

    url = 'https://api.github.com/teams/%s' % team_id
    response = requests.delete(url, auth=github_creds)

    if response.status_code == 204:
        logging.info('Successfully deleted GitHub team called %s with ID %s' % (name, team_id))
        return {}
    else:
        raise Exception('Failed to delete team called %s. Got response %d from GitHub with body: %s.' % (name, response.status_code, response.json()))


def dasherize(name):
    """
    Convert the given name to a lower case, dash-separated string. E.g., "Foo Bar" becomes "foo-bar".
    :param name: The name to dasherize
    :return: The dasherized version of name.
    """
    return re.sub(r'\s', '-', name).lower()


def create_github_team_if_necessary(company_name, subscription_type, github_creds):
    """
    Create the GitHub team for the given company and subscription type, unless the team already exists, in which case,
    raise an Exception.
    :param company_name: The name of the company
    :param subscription_type: The subscription type
    :param github_creds: The GitHub creds to use for the API call. Should be a tuple of (username, password).
    :return: The return value of create_github_team
    """
    team_name = dasherize(company_name)
    team_description = 'Gruntwork customer %s' % company_name
    team_repos = repos_for_subscription[subscription_type]

    if find_github_team(team_name, github_creds):
        raise Exception('Team %s already exists! Cannot create again.' % team_name)
    else:
        return create_github_team(team_name, team_description, team_repos, github_creds)


def remove_github_team_if_necessary(company_name, github_creds):
    """
    Remove the GitHub team for the given company and subscription type, if it already exists. If it doesn't, raise an
    Exception.
    :param company_name: The company name
    :param github_creds: The GitHub creds to use for the API call. Should be a tuple of (username, password).
    :return: The return value of remove_github_team
    """
    team_name = dasherize(company_name)
    team_id = find_github_team(team_name, github_creds)

    if not team_id:
        raise Exception('Did not find a GitHub team called %s.' % team_name)

    return remove_github_team(team_name, team_id, github_creds)


def run():
    """
    Main entrypoint for the code. Reads data from the environment and creates the GitHub team. Returns the response body
    of the GitHub create team API call.
    """
    github_user = read_from_env('GITHUB_USER')
    github_pass = read_from_env('GITHUB_TOKEN')

    assert len(github_user) > 2, 'GitHub username does not seem to be valid (less than 3 characters long)'
    assert len(github_pass) > 2, 'GitHub password does not seem to be valid (less than 3 characters long)'

    github_creds = (github_user, github_pass)

    company_name = read_from_env('company_name')
    subscription_type = read_from_env('subscription_type')
    active = read_from_env('active', required=False)

    assert len(company_name) > 2, 'Company name does not seem to be valid (less than 3 characters long)'
    assert subscription_type in ['aws', 'gcp', 'enterprise'], 'Invalid subscription type. Must be one of: aws, gcp, enterprise.'

    if active == "Yes":
        logging.info('The "active" input is set to "Yes", so creating a new GitHub team for company %s.' % company_name)
        return create_github_team_if_necessary(company_name, subscription_type, github_creds)
    elif active == "No":
        logging.info('The "active" input is set to "No", so deleting the GitHub team for company %s.' % company_name)
        return remove_github_team_if_necessary(company_name, github_creds)
    else:
        logging.info('The "active" input is not set to "Yes" or "No", so assuming this entry is still a WIP and will not take any action.')
        return {}


# Zapier requires that you set a variable called output with your returned data
logging.basicConfig(format='%(asctime)s [%(levelname)s] %(message)s', level=logging.INFO)
output = run()
logging.info(output)
