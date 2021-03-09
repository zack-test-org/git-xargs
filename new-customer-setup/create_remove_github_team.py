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
    'gruntwork-io/fetch-test-private',
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
    'gruntwork-io/terraform-aws-asg',
    'gruntwork-io/terraform-aws-monitoring',
    'gruntwork-io/terraform-aws-cache',
    'gruntwork-io/terraform-aws-ci',
    'gruntwork-io/terraform-aws-ci-pipeline-example',
    'gruntwork-io/terraform-aws-data-storage',
    'gruntwork-io/terraform-aws-ecs',
    'gruntwork-io/terraform-aws-load-balancer',
    'gruntwork-io/terraform-aws-security',
    'gruntwork-io/terraform-aws-server',
    'gruntwork-io/terraform-aws-vpc',
    'gruntwork-io/terraform-aws-beanstalk',
    'gruntwork-io/terraform-aws-elk',
    'gruntwork-io/terraform-aws-kafka',
    'gruntwork-io/terraform-aws-lambda',
    'gruntwork-io/terraform-aws-messaging',
    'gruntwork-io/terraform-aws-openvpn',
    'gruntwork-io/terraform-aws-sam',
    'gruntwork-io/terraform-aws-static-assets',
    'gruntwork-io/terraform-aws-terraform-utilities',
    'gruntwork-io/terraform-aws-zookeeper',
    'gruntwork-io/terraform-aws-mongodb',
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
    'hashicorp/terraform-aws-consul',
    'hashicorp/terraform-aws-nomad',
    'hashicorp/terraform-aws-vault',
    'gruntwork-io/terraform-aws-service-catalog',
    'gruntwork-io/aws-sample-app',
    'gruntwork-io/circle-ci-docker-images'
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
    'gruntwork-io/terraform-aws-ci',
    'gruntwork-io/terraform-aws-ci-pipeline-example',
    'gruntwork-io/terraform-aws-security',
    'gruntwork-io/terraform-kubernetes-helm',
    'gruntwork-io/terraform-helm-gke-exts',
    'gruntwork-io/terraform-google-security',
    'gruntwork-io/terratest',
    'gruntwork-io/terragrunt',
    'gruntwork-io/terragrunt-infrastructure-live-example',
    'gruntwork-io/terragrunt-infrastructure-modules-example',
    'gruntwork-io/toc',
    'hashicorp/terraform-google-consul',
    'hashicorp/terraform-google-nomad',
    'hashicorp/terraform-google-vault',
    'gruntwork-io/circle-ci-docker-images'
]

aws_cis_repos = [
    'gruntwork-io/terraform-aws-cis-service-catalog',
    'gruntwork-io/cis-infrastructure-live-acme',
    'gruntwork-io/cis-infrastructure-modules-acme',
]

repos_for_subscription = {
    'aws': aws_repos,
    'gcp': gcp_repos,
    'enterprise': list(set(aws_repos + gcp_repos)),
    'aws-cis': list(set(aws_repos + aws_cis_repos)),
    'enterprise-cis': list(set(aws_repos + gcp_repos + aws_cis_repos))
}

# Regex to match "next" URLs in the GitHub API Link header. Example: <https://api.github.com/foo/bar?page=2>; rel="next"
github_api_next_regex = re.compile('<(.+?)>; rel="next"')


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
        raise Exception('Did not find value for key {} in either input_data or environment variables.'.format(key))

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


def gen_github_repo_from_template(repo_name, github_creds):
    """
    Generate a new repository from an existing template repo.
    :param repo_name: The name to use for the new repo.
    :param github_creds: The GitHub creds to use for the API call. Should be a tuple of (username, password).
    :return: An empty object.
    """
    description = 'This repository is used for deploying the Gruntwork Reference Architecture.'
    headers = {
        # This API is in preview. See: https://developer.github.com/changes/2019-07-16-repository-templates-api/
        'Accept': 'application/vnd.github.baptiste-preview+json'
    }
    payload = {
        'name': repo_name,
        'description': description,
        'private': True,
        'owner': 'gruntwork-clients'
    }

    api_url = 'https://api.github.com/repos/gruntwork-clients/infrastructure-live/generate'
    response = requests.post(api_url, auth=github_creds, headers=headers, json=payload)

    if response.status_code == 201:
        response_body = response.json()
        repo_id = response_body['id']
        logging.info('Successfully created GitHub repo called {} with ID {}'.format(repo_name, repo_id))
        return {}

    raise Exception('Failed to create repo called {}. Got response {} from GitHub with body: {}.'.format(repo_name, response.status_code, response.json()))


def find_github_team_slug(name, github_creds, teams_api_url='', github_org='gruntwork-io'):
    """
    Find the slug of a GitHub team with the given name. Returns the team slug or None.
    :param name: The name of the GitHub team
    :param github_creds: The GitHub creds to use for the API call. Should be a tuple of (username, password).
    :param teams_api_url: The GitHub API URL to use. Used solely by this function to make recursive API calls
      paginating.
    :param github_org: The GitHub org in which to search.
    :return: The slug of the team or None
    """
    team = find_github_team(name, github_creds, github_org=github_org)
    if team:
        return team['slug']

    logging.info('No team with name {} found'.format(name))
    return None


def find_github_team_id(name, github_creds, teams_api_url='', github_org='gruntwork-io'):
    """
    Find the ID of a GitHub team with the given name. Returns the team ID or None.
    :param name: The name of the GitHub team
    :param github_creds: The GitHub creds to use for the API call. Should be a tuple of (username, password).
    :param teams_api_url: The GitHub API URL to use. Used solely by this function to make recursive API calls
      paginating.
    :param github_org: The GitHub org in which to search.
    :return: The ID of the team or None
    """
    team = find_github_team(name, github_creds, github_org=github_org)
    if team:
        return team['id']

    logging.info('No team with name {} found'.format(name))
    return None


def find_github_team(name, github_creds, teams_api_url='', github_org='gruntwork-io'):
    """
    Find a GitHub team with the given name. Returns the team or None. Note that we have to use the teams list API
    (https://developer.github.com/v3/teams/#list-teams) to find the team because the API call to fetch a team by name
    (https://developer.github.com/v3/teams/#get-team-by-name) requires a "slug", which is not the original team name
    you used, but a version of that name with special characters modified in a variety of ways (e.g., whitespace
    replaced with dashes, dots replaced with dashes, ampersands and surrounding whitespace dropped, etc.) that don't
    seem to be published publicly.
    :param name: The name of the GitHub team
    :param github_creds: The GitHub creds to use for the API call. Should be a tuple of (username, password).
    :param teams_api_url: The GitHub API URL to use. Used solely by this function to make recursive API calls
      paginating.
    :param github_org: The GitHub org in which to search.
    :return: The ID of the team or None
    """
    if teams_api_url == '':
        teams_api_url = 'https://api.github.com/orgs/' + github_org + '/teams?per_page=100'

    logging.info('Looking up team {} in GitHub API via URL {}...'.format(name, teams_api_url))
    response = requests.get(teams_api_url, auth=github_creds)
    if response.status_code != 200:
        raise Exception('Got response {} from GitHub when searching for teams at URL {}'.format(response.status_code, teams_api_url))

    teams = response.json()
    for team in teams:
        if team['name'] == name:
            logging.info('Found ID {} for GitHub team {}'.format(team['id'], name))
            return team

    next_url = response.links.get('next')
    if next_url:
        return find_github_team(name, github_creds, next_url['url'])

    logging.info('No team with name {} found'.format(name))
    return None


def create_github_team(company_name, repos, github_creds, github_org='gruntwork-io'):
    """
    Create a GitHub team in the given org with the given name and description. https://developer.github.com/v3/teams/#create-team
    :param company_name: The name of the the customer company
    :param repos: The repos the team should get access to
    :param github_creds: The GitHub creds to use for the API call. Should be a tuple of (username, password).
    :param github_org: The GitHub org in which to create the team (e.g. gruntwork-io)
    :return: Returns the JSON body of the GitHub create-team API response
    """
    team_name = format_github_team_name(company_name)
    logging.info('Creating new GitHub team called {} and granting it access to repos: {}'.format(team_name, repos))
    team_description = 'Gruntwork customer {}'.format(company_name)

    payload = {
        'name': team_name,
        'description': team_description,
        'privacy': 'secret',
        'repo_names': repos
    }

    api_url = 'https://api.github.com/orgs/' + github_org + '/teams'
    response = requests.post(api_url, auth=github_creds, json=payload)

    if response.status_code == 201:
        response_body = response.json()
        team_id = response_body['id']
        logging.info('Successfully created GitHub team called {} with ID {}'.format(team_name, team_id))
        return response_body

    raise Exception('Failed to create team called {}. Got response {} from GitHub with body: {}.'.format(team_name, response.status_code, response.json()))


def remove_github_team(name, team_id, github_creds):
    """
    Remove a GitHub team with the given name and description.
    https://developer.github.com/v3/teams/#delete-team
    :param name: The name of the team to remove. Solely used for logging.
    :param team_id: The ID of the team to remove (NOT the name)
    :param github_creds: The GitHub creds to use for the API call. Should be a tuple of (username, password).
    :return: An empty object
    """
    logging.info('Deleting GitHub team called {} with ID {}'.format(name, team_id))

    url = 'https://api.github.com/teams/{}'.format(team_id)
    response = requests.delete(url, auth=github_creds)

    if response.status_code == 204:
        logging.info('Successfully deleted GitHub team called {} with ID {}'.format(name, team_id))
        return {}

    raise Exception('Failed to delete team called {}. Got response {} from GitHub with body: {}.'.format(name, response.status_code, response.json()))


def format_company_name(name):
    """
    Convert the given name to a GitHub team name for a customer. We do this by converting the name to a lower case,
    dash-separated string. E.g., "Foo Bar.com" becomes "foo-bar-com".
    :param name: The name to dasherize
    :return: The friendly version of name.
    """
    return '{}'.format(re.sub(r'[\s.]', '-', name).lower())


def format_github_team_name(name):
    """
    Convert the given name to a GitHub team name for a customer. E.g., "Foo Bar.com" becomes "client-foo-bar-com".
    :param name: The name to dasherize
    :return: The GitHub-friendly team version of name.
    """
    return 'client-' + format_company_name(name)


def format_infra_live_repo_name(name):
    """
    Convert the given name to a customer infrastructure-live repo name. E.g. "Foo Bar.com" becomes
    "infrastructure-live-foo-bar-com".
    :param name: The name to dasherize
    :return: The infrastructure-live repo name.
    """
    return 'infrastructure-live-' + format_company_name(name)


def create_github_team_if_necessary(company_name, subscription_type, github_creds):
    """
    Create the GitHub team for the given company and subscription type, unless the team already exists, in which case,
    raise an Exception.
    :param company_name: The name of the company
    :param subscription_type: The subscription type
    :param github_creds: The GitHub creds to use for the API call. Should be a tuple of (username, password).
    :return: The return value of create_github_team
    """
    team_repos = repos_for_subscription[subscription_type]
    team_name = format_github_team_name(company_name)

    if find_github_team_id(team_name, github_creds):
        raise Exception('Team {} already exists! Cannot create again.'.format(team_name))

    return create_github_team(company_name, team_repos, github_creds)


def remove_github_team_if_necessary(company_name, github_creds, ref_arch):
    """
    Remove the GitHub team for the given company and subscription type, if it already exists. If it doesn't, raise an
    Exception.
    :param company_name: The company name
    :param github_creds: The GitHub creds to use for the API call. Should be a tuple of (username, password).
    :param ref_arch: Boolean indicating whether or not the customer purchased a Gruntwork Ref Arch.
    :return: The return value of remove_github_team
    """
    team_name = format_github_team_name(company_name)
    if ref_arch:
        clients_team_id = find_github_team_id(team_name, github_creds, 'gruntwork-clients')
        remove_github_team(team_name, clients_team_id, github_creds)

    team_id = find_github_team_id(team_name, github_creds)

    if not team_id:
        raise Exception('Did not find a GitHub team called {}.'.format(team_name))

    return remove_github_team(team_name, team_id, github_creds)


def update_template_repo_permissions(team_name, repo_name, github_org, github_creds):
    """
    Update the repository permissions to grant write access to the given team.
    :param team_name: The name of the GitHub team
    :param repo_name: The name of the repository for granting write access.
    :param github_org: The GitHub org where the repo is located
    :param github_creds: The GitHub creds to use for the API call. Should be a tuple of (username, password).
    :return: An empty object
    """
    clients_team_slug = find_github_team_slug(team_name, github_creds, github_org=github_org)
    api_url = 'https://api.github.com/orgs/{}/teams/{}/repos/{}/{}'.format(github_org, clients_team_slug, github_org, repo_name)
    payload = {
        'permission': 'push'
    }
    response = requests.put(api_url, auth=github_creds, json=payload)

    if response.status_code != 204:
        raise Exception('Failed with response {} body: {}.'.format(response.status_code, response.json()))


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
    Main entrypoint for the code. Reads data from the environment and creates the GitHub team. Returns the response body
    of the GitHub create team API call.
    """
    github_user = read_from_env('GITHUB_USER')
    github_pass = read_from_env('GITHUB_TOKEN')
    github_clients_user = read_from_env('GITHUB_CLIENTS_USER')
    github_clients_pass = read_from_env('GITHUB_CLIENTS_TOKEN')

    assert len(github_user) > 2, 'GitHub username does not seem to be valid (less than 3 characters long)'
    assert len(github_pass) > 2, 'GitHub password does not seem to be valid (less than 3 characters long)'
    assert len(github_clients_user) > 2, 'GitHub clients username does not seem to be valid (less than 3 characters long)'
    assert len(github_clients_pass) > 2, 'GitHub clients password does not seem to be valid (less than 3 characters long)'

    github_creds = (github_user, github_pass)
    github_clients_creds = (github_clients_user, github_clients_pass)

    company_name = read_from_env('company_name')
    subscription_type = read_from_env('subscription_type')
    active = read_from_env('active', required=False)
    ref_arch = read_from_env('ref_arch')

    assert len(company_name) > 1, 'Company name does not seem to be valid (less than 2 characters long)'
    assert subscription_type in repos_for_subscription, 'Invalid subscription type. Must be one of: {}'.format(list(repos_for_subscription.keys()))

    if is_affirmative_value(active) and is_affirmative_value(ref_arch):
        team_name = format_github_team_name(company_name)
        if find_github_team_id(team_name, github_clients_creds, github_org='gruntwork-clients'):
            logging.info('Found GitHub team for gruntwork-clients, not creating')
        else:
            infra_live_repo = format_infra_live_repo_name(company_name)
            gen_github_repo_from_template(infra_live_repo, github_clients_creds)
            create_github_team(company_name, ['gruntwork-clients/' + infra_live_repo], github_clients_creds, 'gruntwork-clients')
            update_template_repo_permissions(team_name, infra_live_repo, 'gruntwork-clients', github_clients_creds)
            logging.info('Created GitHub team in gruntwork-clients for customer {}'.format(company_name))

    if is_affirmative_value(active):
        logging.info('The "active" input is set to "Yes", so creating a new GitHub team for company {}.'.format(company_name))
        return create_github_team_if_necessary(company_name, subscription_type, github_creds)
    elif is_negative_value(active):
        logging.info('The "active" input is set to "No", so deleting the GitHub team for company {}.'.format(company_name))
        return remove_github_team_if_necessary(company_name, github_creds, is_affirmative_value(ref_arch))
    else:
        logging.info('The "active" input is not set to "Yes" or "No", so assuming this entry is still a WIP and will not take any action.')
        return {}


# Zapier requires that you set a variable called output with your returned data
logging.basicConfig(format='%(asctime)s [%(levelname)s] %(message)s', level=logging.INFO)
output = run()
logging.info(output)
