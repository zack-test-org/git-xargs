import urllib.request
import urllib.parse
import json
import os
import re

github_token = os.environ['GITHUB_OAUTH_TOKEN']


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


def make_github_request_all_pages(path=None, url=None, body_so_far=[]):
    body, headers = make_github_request(path=path, url=url)
    body_so_far += body

    next_link = get_next_link(headers)
    if next_link:
        return make_github_request_all_pages(path=None, url=next_link, body_so_far=body_so_far)
    else:
        return body_so_far


github_api_next_regex = re.compile('<(.+?)>; rel="next"')


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
    # For some reason, the collaborators API returns not only users, but also other repos... And those repos don't have
    # a 'login' field... So we ignore them here.
    return collaborator.get('permissions', {}).get('admin', False) and collaborator.get('login', False)


def get_users_with_admin_permissions(repo):
    collaborators = make_github_request_all_pages(f'/repos/gruntwork-io/{repo}/collaborators?per_page=100')
    return [collaborator['login'] for collaborator in collaborators if is_collaborator_admin(collaborator)]


def get_all_repos():
    return make_github_request_all_pages('/orgs/gruntwork-io/repos?per_page=100')


if __name__ == '__main__':
    print('Fetching list of repos in gruntwork-io org...')
    repos = get_all_repos()
    print(f'Found {len(repos)} repositories.')

    unexpected_repo_admins = {}
    for repo in repos:
        repo_name = repo['name']
        print(f'Looking up users for repo {repo_name}...')
        admins_for_repo = get_users_with_admin_permissions(repo_name)
        print(f'Admins for {repo_name}: {set(admins_for_repo)}')

        unexpected_admins_for_repo = [admin for admin in admins_for_repo if admin not in ['brikis98', 'josh-padnick']]
        if unexpected_admins_for_repo:
            unexpected_repo_admins[repo_name] = unexpected_admins_for_repo

    print('\n\n======= RESULTS =======\n')

    if unexpected_repo_admins:
        print('Unexpected admins found!')
        for repo, admins in unexpected_repo_admins:
            print(f'{repo}: {admins}')
    else:
        print('No unexpected admins found!')
