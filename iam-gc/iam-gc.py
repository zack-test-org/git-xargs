import boto3
import re
import datetime as dt
import pytz
from tabulate import tabulate
from progress.bar import Bar


def get_all_roles():
    client = boto3.client('iam')
    response = client.list_roles()
    roles = response['Roles']
    while 'Marker' in response:
        response = client.list_roles(Marker=response['Marker'])
        roles.extend(response['Roles'])
    return roles


def delete_role(role):
    client = boto3.client('iam')
    name = role['RoleName']
    create_date = role['CreateDate']
    print('Deleting role {} created on {}'.format(name, create_date.isoformat()))
    client.delete_role(RoleName=name)


def get_all_profiles():
    client = boto3.client('iam')
    response = client.list_instance_profiles()
    profiles = response['InstanceProfiles']
    while 'Marker' in response:
        response = client.list_instance_profiles(Marker=response['Marker'])
        profiles.extend(response['InstanceProfiles'])
    return profiles


def delete_profile(profile):
    client = boto3.client('iam')
    name = profile['InstanceProfileName']
    create_date = profile['CreateDate']
    print('Deleting profile {} created on {}'.format(name, create_date.isoformat()))
    client.delete_instance_profile(InstanceProfileName=name)


def is_test_role_or_instance_profile(name):
    regex_list = [
        r'^cloud-nuke-test-.+',
        r'^cloud-nuke-Test.+',
        r'^vault-test-[a-zA-Z0-9]{6}.+',
        r'^consul-test-[a-zA-Z0-9]{6}.+',
        r'^kibana-[a-zA-Z0-9]{6}.+',
        r'^server-group.+',
        r'^[Tt]est-cluster[a-zA-Z0-9]{6}$',
        r'^[a-zA-Z0-9]{6}-cluster$',
        r'^[a-zA-Z0-9]{6}-ecs-cluster$',
    ]
    return any(re.match(regex, name) for regex in regex_list)


def created_before(t):
    unow = dt.datetime.utcnow().replace(tzinfo=pytz.utc)
    return t < unow - dt.timedelta(hours=24)


def want_profile(profile):
    name = profile['InstanceProfileName']
    create_date = profile['CreateDate']
    return is_test_role_or_instance_profile(name) and created_before(create_date)


def want_role(role):
    name = role['RoleName']
    create_date = role['CreateDate']
    return is_test_role_or_instance_profile(name) and created_before(create_date)


def run_roles(dry=True):
    roles = get_all_roles()
    roles = list(filter(want_role, roles))
    print('Found {} roles'.format(len(roles)))
    print('Last role created {}'.format(max(role['CreateDate'] for role in roles).isoformat()))
    roles.sort(key=lambda p: p['CreateDate'], reverse=True)
    print(
        tabulate(
            [(role['RoleName'], role['CreateDate'].isoformat()) for role in roles],
            headers=('Name', 'Created'),
        )
    )
    print()

    if dry:
        return

    input('Will delete {} roles. [Ctrl+C] to cancel, or [ENTER] to proceed.'.format(len(roles)))
    bar = Bar('Deleting', max=len(roles))
    for role in bar.iter(roles):
        delete_role(role)


def run_profiles(dry=True):
    profiles = get_all_profiles()
    profiles = list(filter(want_profile, profiles))
    print('Found {} profiles'.format(len(profiles)))
    print('Last profile created {}'.format(max(profile['CreateDate'] for profile in profiles).isoformat()))
    profiles.sort(key=lambda p: p['CreateDate'], reverse=True)
    print(
        tabulate(
            [(profile['InstanceProfileName'], profile['CreateDate'].isoformat()) for profile in profiles],
            headers=('Name', 'Created'),
        )
    )
    print()

    if dry:
        return

    input('Will delete {} profiles. [Ctrl+C] to cancel, or [ENTER] to proceed.'.format(len(profiles)))
    bar = Bar('Deleting', max=len(profiles))
    for profile in bar.iter(profiles):
        delete_profile(profile)


def run(dry=True):
    run_profiles(dry)
    run_roles(dry)


if __name__ == '__main__':
    run(dry=True)
