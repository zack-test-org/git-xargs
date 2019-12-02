import requests
import logging
import os

# TODO: plug in creds for machine user
github_user = 'brikis98'
github_pass = os.environ['GITHUB_TOKEN']
github_creds = (github_user, github_pass)

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
    {'owner': 'gruntwork-io', 'name': "infrastructure-live-google"},
    {'owner': 'gruntwork-io', 'name': "infrastructure-modules-google"},
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
def create_github_team(name, description, repos):
    logging.info('Creating new GitHub team called %s' % name)

    payload = {
        'name': name,
        'description': description,
        'privacy': 'secret',
        'repo_names': repos
    }

    response = requests.post('https://api.github.com/orgs/gruntwork-io/teams', auth=github_creds, json=payload)

    if response.status_code == 201:
        team_id = response.json()['id']
        logging.info('Successfully created GitHub team called %s with ID %s' % (name, team_id))
        return team_id
    else:
        raise Exception('Failed to create team called %s. Got response %d from GitHub with body: %s.' % (name, response.status_code, response.json()))


def run():
    logging.basicConfig(format='%(asctime)s [%(levelname)s] %(message)s', level=logging.INFO)

    # TODO: need name and list of repos to be passed in dynamically
    name = 'jim-testing'
    description = 'Jim testing the GitHub APIs and Zapier'
    repos = aws_repos

    if find_github_team(name):
        logging.info('Team %s already exists. Will not create again.' % name)
    else:
        create_github_team(name, description, repos)


if __name__ == '__main__':
    run()
