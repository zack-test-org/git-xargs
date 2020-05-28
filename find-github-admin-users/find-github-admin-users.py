import urllib.request
import urllib.parse
import json
import os
import re

github_token = os.environ['GITHUB_OAUTH_TOKEN']
github_api_next_regex = re.compile('<(.+?)>; rel="next"')
expected_admins = {'brikis98', 'josh-padnick'}


def make_github_request(path=None, url=None, expected_status=200):
    if not url:
        url = urllib.parse.urljoin('https://api.github.com/', path)

    print(f'Making API call to {url}')

    headers = {
        'Authorization': f'token {github_token}'
    }

    req = urllib.request.Request(url, headers=headers)

    response = urllib.request.urlopen(req)

    status_code = response.getcode()
    body = response.read()
    headers = response.getheaders()

    if status_code != expected_status:
        raise Exception(f'Expected status {expected_status} but got {status_code}. Body: {body}')

    json_body = json.loads(body)

    return json_body, headers


def make_github_request_all_pages(path=None, url=None, body_so_far=None):
    body, headers = make_github_request(path=path, url=url)
    result = body_so_far + body if body_so_far else body

    next_link = get_next_link(headers)
    if next_link:
        return make_github_request_all_pages(path=None, url=next_link, body_so_far=result)
    else:
        return result


def get_next_link(headers):
    for header in headers:
        if header[0] == "Link":
            links = header[1].split(', ')
            for link in links:
                matches = github_api_next_regex.search(link)
                if matches:
                    return matches.group(1)
    return None


def is_collaborator_admin(collaborator):
    return collaborator.get('permissions', {}).get('admin', False)


def get_users_with_admin_permissions(repo):
    collaborators = make_github_request_all_pages(f'/repos/gruntwork-io/{repo}/collaborators?per_page=100')
    return [collaborator['login'] for collaborator in collaborators if is_collaborator_admin(collaborator)]


def get_all_repos():
    print('Fetching list of repos in gruntwork-io org...')
    all_repos = make_github_request_all_pages('/orgs/gruntwork-io/repos?per_page=100')
    print(f'Found {len(all_repos)} repositories.')
    return all_repos


def get_all_admins(repos):
    all_admins = {}
    hit_errors = False

    for index, repo in enumerate(repos):
        try:
            repo_name = repo['name']
            print(f'Processing repo {index + 1} / {len(repos)}: Looking up users for repo {repo_name}...')
            admins_for_repo = set(get_users_with_admin_permissions(repo_name))
            print(f'Admins for {repo_name}: {admins_for_repo}')
            all_admins[repo_name] = admins_for_repo
        except Exception as err:
            print(f'Caught an error while processing repo {repo}: {err}')
            hit_errors = True

    return all_admins, hit_errors


def get_unexpected_admins(all_admins):
    all_unexpected_admins = {}

    for repo, admins in all_admins.items():
        unexpected_admins = admins - expected_admins
        if unexpected_admins:
            all_unexpected_admins[repo] = unexpected_admins

    return all_unexpected_admins


def print_results(all_unexpected_admins, hit_errors):
    print('\n\n======= RESULTS =======\n')

    if all_unexpected_admins:
        print('Unexpected admins found!')
        for repo, admins in all_unexpected_admins.items():
            print(f'{repo}: {admins}')
    else:
        print('No unexpected admins found!')

    if hit_errors:
        print('\nWARNING: There were errors during the search. See the logs above!')


def print_unexpected_admins():
    all_repos = get_all_repos()
    all_admins, hit_errors = get_all_admins(all_repos)
    all_unexpected_admins = get_unexpected_admins(all_admins)
    print_results(all_unexpected_admins, hit_errors)


if __name__ == '__main__':
    print_unexpected_admins()
