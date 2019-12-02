import requests
import logging
import os

# TODO: plug in creds for machine user
github_user = 'brikis98'
github_pass = os.environ['GITHUB_TOKEN']
github_creds = (github_user, github_pass)

aws_repos = [
    {'owner': 'gruntwork-io', 'name': 'bash-commons'},
    {'owner': 'gruntwork-io', 'name': 'cloud-nuke'},
    {'owner': 'gruntwork-io', 'name': 'fetch'},
    {'owner': 'gruntwork-io', 'name': 'gruntkms'},
    {'owner': 'gruntwork-io', 'name': 'gruntwork'},
    {'owner': 'gruntwork-io', 'name': 'gruntwork-cli'},
    {'owner': 'gruntwork-io', 'name': 'gruntwork-installer'},
    {'owner': 'gruntwork-io', 'name': 'helm-kubernetes-services'},
    {'owner': 'gruntwork-io', 'name': 'infrastructure-as-code-training'},
    {'owner': 'gruntwork-io', 'name': 'infrastructure-live-acme'},
    {'owner': 'gruntwork-io', 'name': 'infrastructure-live-multi-account-acme'},
    {'owner': 'gruntwork-io', 'name': 'infrastructure-modules-acme'},
    {'owner': 'gruntwork-io', 'name': 'infrastructure-modules-multi-account-acme'},
    {'owner': 'gruntwork-io', 'name': 'intro-to-terraform'},
    {'owner': 'gruntwork-io', 'name': 'kubergrunt'},
    {'owner': 'gruntwork-io', 'name': 'module-asg'},
    {'owner': 'gruntwork-io', 'name': 'module-aws-monitoring'},
    {'owner': 'gruntwork-io', 'name': 'module-cache'},
    {'owner': 'gruntwork-io', 'name': 'module-ci'},
    {'owner': 'gruntwork-io', 'name': 'module-data-storage'},
    {'owner': 'gruntwork-io', 'name': 'module-ecs'},
    {'owner': 'gruntwork-io', 'name': 'module-load-balancer'},
    {'owner': 'gruntwork-io', 'name': 'module-security'},
    {'owner': 'gruntwork-io', 'name': 'module-server'},
    {'owner': 'gruntwork-io', 'name': 'module-vpc'},
    {'owner': 'gruntwork-io', 'name': 'package-beanstalk'},
    {'owner': 'gruntwork-io', 'name': 'package-elk'},
    {'owner': 'gruntwork-io', 'name': 'package-kafka'},
    {'owner': 'gruntwork-io', 'name': 'package-lambda'},
    {'owner': 'gruntwork-io', 'name': 'package-messaging'},
    {'owner': 'gruntwork-io', 'name': 'package-openvpn'},
    {'owner': 'gruntwork-io', 'name': 'package-sam'},
    {'owner': 'gruntwork-io', 'name': 'package-static-assets'},
    {'owner': 'gruntwork-io', 'name': 'package-terraform-utilities'},
    {'owner': 'gruntwork-io', 'name': 'package-zookeeper'},
    {'owner': 'gruntwork-io', 'name': 'package-mongodb'},
    {'owner': 'gruntwork-io', 'name': 'sample-app-backend-acme'},
    {'owner': 'gruntwork-io', 'name': 'sample-app-backend-multi-account-acme'},
    {'owner': 'gruntwork-io', 'name': 'sample-app-backend-packer'},
    {'owner': 'gruntwork-io', 'name': 'sample-app-frontend-acme'},
    {'owner': 'gruntwork-io', 'name': 'sample-app-frontend-multi-account-acme'},
    {'owner': 'gruntwork-io', 'name': 'terraform-aws-couchbase'},
    {'owner': 'gruntwork-io', 'name': 'terraform-aws-eks'},
    {'owner': 'gruntwork-io', 'name': 'terraform-kubernetes-helm'},
    {'owner': 'gruntwork-io', 'name': 'terratest'},
    {'owner': 'gruntwork-io', 'name': 'terragrunt'},
    {'owner': 'gruntwork-io', 'name': 'terragrunt-infrastructure-live-example'},
    {'owner': 'gruntwork-io', 'name': 'terragrunt-infrastructure-modules-example'},
    {'owner': 'gruntwork-io', 'name': 'toc'},
]

gcp_repos = [
    {'owner': 'gruntwork-io', 'name': 'bash-commons'},
    {'owner': 'gruntwork-io', 'name': 'cloud-nuke'},
    {'owner': 'gruntwork-io', 'name': 'fetch'},
    {'owner': 'gruntwork-io', 'name': 'gruntwork-cli'},
    {'owner': 'gruntwork-io', 'name': 'gruntwork-installer'},
    {'owner': 'gruntwork-io', 'name': 'helm-kubernetes-services'},
    {'owner': 'gruntwork-io', 'name': 'infrastructure-as-code-training'},
    {'owner': 'gruntwork-io', 'name': "infrastructure-live-google"},
    {'owner': 'gruntwork-io', 'name': "infrastructure-modules-google"},
    {'owner': 'gruntwork-io', 'name': 'intro-to-terraform'},
    {'owner': 'gruntwork-io', 'name': 'kubergrunt'},
    {'owner': 'gruntwork-io', 'name': 'module-ci'},
    {'owner': 'gruntwork-io', 'name': 'module-security'},
    {'owner': 'gruntwork-io', 'name': 'terraform-kubernetes-helm'},
    {'owner': 'gruntwork-io', 'name': 'terratest'},
    {'owner': 'gruntwork-io', 'name': 'terragrunt'},
    {'owner': 'gruntwork-io', 'name': 'terragrunt-infrastructure-live-example'},
    {'owner': 'gruntwork-io', 'name': 'terragrunt-infrastructure-modules-example'},
    {'owner': 'gruntwork-io', 'name': 'toc'},
]


# Find a GitHub team with the given name. Returns the team ID or None.
# https://developer.github.com/v3/teams/#get-team
def find_github_team(name):
    logging.info('Looking for a GitHub team called %s' % name)
    response = requests.get('https://api.github.com/orgs/gruntwork-io/teams/%s' % name, auth=github_creds)
    if response.status_code == 200:
        team_id = response.json()['id']
        logging.info('Found GitHub team with ID %s' % team_id)
        return team_id
    else:
        logging.info('No team with name %s found (got response %d from GitHub)' % (name, response.status_code))
        return None


# Create a GitHub team with the given name and description. Returns the team ID.
# https://developer.github.com/v3/teams/#create-team
def create_github_team(name, description):
    logging.info('Creating new GitHub team called %s' % name)

    payload = {
        'name': name,
        'description': description,
        'privacy': 'secret'
    }

    response = requests.post('https://api.github.com/orgs/gruntwork-io/teams', auth=github_creds, json=payload)

    if response.status_code == 201:
        team_id = response.json()['id']
        logging.info('Successfully created GitHub team called %s with ID %s' % (name, team_id))
        return team_id
    else:
        raise Exception('Failed to create team called %s. Got response %d from GitHub with body: %s.' % (name, response.status_code, response.json()))


# Add the given repo to the GitHub team with the given ID.
# https://developer.github.com/v3/teams/#add-or-update-team-repository
def add_repo_to_team(repo_owner, repo_name, team_id):
    logging.info("Adding repo %s/%s to team %s" % (repo_owner, repo_name, team_id))

    url = 'https://api.github.com/teams/%s/repos/%s/%s' % (team_id, repo_owner, repo_name)
    payload = {
        'permission': 'pull'
    }

    response = requests.put(url, auth=github_creds, json=payload)
    if response.status_code != 204:
        raise Exception('Failed to add repo %s/%s to team %s. Got response code %d with body: %s.' % (repo_owner, repo_name, team_id, response.status_code, response.json()))


def run():
    logging.basicConfig(format='%(asctime)s [%(levelname)s] %(message)s', level=logging.INFO)

    # TODO: need name to be passed in dynamically
    name = 'jim-testing'
    description = 'Jim testing the GitHub APIs and Zapier'

    team_id = find_github_team(name)
    if team_id is None:
        team_id = create_github_team(name, description)

    for repo in aws_repos:
        add_repo_to_team(repo['owner'], repo['name'], team_id)


if __name__ == '__main__':
    run()
