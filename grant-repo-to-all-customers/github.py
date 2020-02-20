import requests
import time
import logging
import gql
from gql.transport.requests import RequestsHTTPTransport
from datetime import datetime

GRUNTWORK_ORG = "gruntwork-io"
MAX_RETRIES = 3


class GithubApiClient:
    """
    A wrapper around gql and requests client that can make requests to the GitHub GraphQL API and GitHub REST API. The
    constructor accepts a single argument, `token`, which should contain a personal access token for authenticating to
    the API.
    """

    _rest_api_base_url = "https://api.github.com"
    _gql_api_url = "https://api.github.com/graphql"

    def __init__(self, token):
        self._api_headers = {'Authorization': f'token {token}'}

        _transport = RequestsHTTPTransport(url=self._gql_api_url, headers=self._api_headers, use_json=True)
        self._gqlclient = gql.Client(transport=_transport)

    def gql_execute(self, query_str, **params):
        query = gql.gql(query_str)
        return self._gqlclient.execute(query, variable_values=params)

    def get(self, url, *args, **kwargs):
        if not url.startswith(self._rest_api_base_url):
            url = self._rest_api_base_url + url
        return self._rate_limit_aware_request(lambda: requests.get(url, *args, **kwargs))

    def put(self, url, *args, **kwargs):
        if not url.startswith(self._rest_api_base_url):
            url = self._rest_api_base_url + url
        return self._rate_limit_aware_request(lambda: requests.put(url, *args, **kwargs))

    def post(self, url, *args, **kwargs):
        if not url.startswith(self._rest_api_base_url):
            url = self._rest_api_base_url + url
        return self._rate_limit_aware_request(lambda: requests.post(url, *args, **kwargs))

    def _rate_limit_aware_request(self, request_func):
        resp = request_func()
        tries = 0
        while _is_rate_limit(resp) and tries < MAX_RETRIES:
            logging.warn('Received rate limit')
            ratelimit_reset_epoch = int(resp.headers['X-RateLimit-Reset'])
            ratelimit_reset_datetime = datetime.fromtimestamp(ratelimit_reset_epoch)
            now = datetime.now()
            diff_seconds = (ratelimit_reset_datetime - now).total_seconds()
            if diff_seconds > 0:
                logging.warn(f'Rate limit resets in {diff_seconds} seconds. Sleeping until reset.')
                time.sleep(diff_seconds)

            logging.warn('Rate limit reset. Retrying request.')
            resp = request_func()
            tries += 1

        resp.raise_for_status()
        return resp


def add_repo_to_team(github_client, team_slug, repo_name):
    """
    Grant access to the repository to the team.

    Args:
        github_client [GithubApiClient] : Configured client to access GitHub API.
        team_slug [str] : The slug of a team in the gruntwork-io org.
        repo_name [str] : The name of a repo in the gruntwork-io org that the team will get access to.
    """
    github_client.put(
        f'/orgs/{GRUNTWORK_ORG}/teams/{team_slug}/repos/{GRUNTWORK_ORG}/{repo_name}',
        json={'permission': 'pull'},
    )


def get_repository_teams(github_client, repo_name):
    """
    Get a list of teams of the Gruntwork org that has access to the given repo.

    Args:
        github_client [GithubApiClient] : Configured client to access GitHub API.
        repo_name [str] : The name of a repo in the gruntwork-io org. All teams that have access to this repo will be
                          returned.

    Returns:
        List of strings representing slugs of teams that have access to the given repo.
    """
    teams = []

    _query = '''query teamsWithRepos($org: String!, $seedRepo: String!, $lastCursor: String) {
      organization(login: $org) {
        teams(first: 100, after: $lastCursor) {
          pageInfo {
            endCursor
            hasNextPage
          }
          nodes {
            slug
            repositories(first: 1, query: $seedRepo) {
              edges {
                node {
                  id
                }
              }
            }
          }
        }
      }
    }'''
    result = github_client.gql_execute(_query, org=GRUNTWORK_ORG, seedRepo=repo_name, lastCursor=None)
    # We make sure to filter out the teams that don't have the seed repo
    teams += [t['slug'] for t in result['organization']['teams']['nodes'] if t['repositories']['edges']]

    # Handle pagination
    while result['organization']['teams']['pageInfo']['hasNextPage']:
        cursor = result['organization']['teams']['pageInfo']['endCursor']
        result = github_client.gql_execute(_query, org=GRUNTWORK_ORG, seedRepo=repo_name, lastCursor=cursor)
        # We make sure to filter out the teams that don't have the seed repo
        teams += [t['slug'] for t in result['organization']['teams']['nodes'] if t['repositories']['edges']]

    return teams


def _is_rate_limit(resp):
    return resp.status_code == 403 and resp.json()['message'].startswith('API rate limit')
