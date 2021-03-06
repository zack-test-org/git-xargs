"""
Set all repositories of a given GitHub organization name for a given user
to watching.

Modified version of https://gist.github.com/thet/c1ce413bdabc771cba1b
"""

from __future__ import print_function

import argparse
import json
import requests
import os
from builtins import input


def get_repos(url, repo_list=None, headers=None):
    if repo_list is None:
        repo_list = []
    if headers is None:
        headers = {}

    resp = requests.get(url, headers=headers)
    resp.raise_for_status()
    repos = resp.json()
    repo_list += repos
    links = getattr(resp, 'links', None)
    if links and 'next' in links and 'url' in links['next']:
        get_repos(links['next']['url'], repo_list=repo_list, headers=headers)
    return repo_list


def main(org_name, dry=False):
    repo_url = 'https://api.github.com/orgs/{0}/repos'.format(org_name)
    access_token = os.environ['GITHUB_OAUTH_TOKEN']
    headers = {'Content-Type': 'application/json; charset=UTF-8', 'Authorization': 'token {}'.format(access_token)}

    repo_list = get_repos(repo_url, headers=headers)

    if dry:
        print('Found {} repos'.format(len(repo_list)))
    else:
        resp = input('Will start watching {} repos. Proceed? [y/n] > '.format(len(repo_list)))
        if resp.strip() != 'y':
            print('Did not respond `y`. Aborting.')
            return
        print('Subscribing to {} repos'.format(len(repo_list)))

    for repo in repo_list:
        repo_name = repo['name']
        print('Subscribing to repo {}'.format(repo_name))
        if dry:
            print('DRY RUN: SKIPPING')
            continue
        url = 'https://api.github.com/repos/{0}/{1}/subscription'.format(
            org_name,
            repo_name
        )
        res = requests.put(
            url=url,
            data='{"subscribed": "1"}',
            headers=headers,
        )
        if res.status_code == 200:
            print('status {0} | repo {1}'.format(
                res.status_code,
                repo_name
            ))
        else:
            print('ERROR! status {0} | repo {1}'.format(
                res.status_code,
                repo_name
            ))


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='Watch/unwatch GitHub Repositories')  # noqa
    parser.add_argument('org_name',  type=str, help='GitHub organization name')
    parser.add_argument('--dry', dest='dry', action='store_true', help='If set, do not actually subscribe and only show logs.')  # noqa
    args = parser.parse_args()
    main(**vars(args))

