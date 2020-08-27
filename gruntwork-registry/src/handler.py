import json
import re
import os
from . import helpers

github_token = helpers.get_github_oauth_token(os.environ['AWS_REGION'], os.environ['GITHUB_OAUTH_TOKEN_SECRETS_MANAGER_ID'])


def get_terraform_service_discovery_json(event, context):
    """
    Implement the Remote Service Discovery endpoint of the Terraform API:
    https://www.terraform.io/docs/internals/remote-service-discovery.html
    :param event:
    :param context:
    :return:
    """
    print(f'event: {event}')
    body = {
        "modules.v1": "/v1/modules/"
    }
    return helpers.format_response(body)


def get_versions_for_module(event, context):
    """
    For a given module in the gruntwork-io org, return the latest versions for it in the format expected by the
    Terraform Registry.
    :param event:
    :param context:
    :return:
    """
    print(f'event: {event}')

    path = event['path']
    match = re.search(r'^/v1/modules/gruntwork-io/(.+?)/(.+?)$', path)

    if not match or len(match.groups()) != 2:
        return helpers.format_response(f'Invalid path: {path}', 400)

    name = match.group(1)
    cloud = match.group(2)

    print(f'name = {name}, cloud = {cloud}')

    # First, try the name as given
    status_code, body = helpers.make_github_request(f'https://api.github.com/repos/gruntwork-io/{name}/releases', github_token)
    if status_code == 404:
        # If that fails, try the Terraform Registry style name
        status_code, body = helpers.make_github_request(
            f'https://api.github.com/repos/gruntwork-io/terraform-{cloud}-{name}/releases', github_token)

    if status_code != 200:
        body_as_str = body.decode('utf-8') if body else None
        return helpers.format_response(body_as_str, status_code)

    releases = json.loads(body)
    versions = [release['tag_name'] for release in releases if not release['draft'] and not release['prerelease']]

    out = {
        "namespace": "gruntwork-io",
        "name": name,
        "provider": cloud,
        "versions": versions
    }

    return helpers.format_response(out)


