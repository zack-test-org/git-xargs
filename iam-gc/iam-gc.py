import boto3
import argparse
import re
import datetime as dt
import pytz
from tabulate import tabulate
from progress.bar import Bar


# ---------------------------------------------------------------------------------------------------------------------
# HELPER FUNCTIONS TO PROCESS IAM USERS
# ---------------------------------------------------------------------------------------------------------------------

def get_all_users():
    """Get all IAM Users in the account, accounting for pagination in the API"""
    return get_all_objects(boto3.client('iam').list_users, 'Users')


def delete_user(user):
    """Delete the given IAM User"""
    client = boto3.client('iam')
    name = user['UserName']
    create_date = user['CreateDate']
    print('Deleting user {} created on {}'.format(name, create_date.isoformat()))
    print('Detaching group from user')
    for group in get_all_groups_for_user(name):
        client.remove_user_from_group(GroupName=group['GroupName'], UserName=name)
    print('Detaching all policies on the user')
    for policy in get_all_user_policies(name):
        client.detach_user_policy(UserName=name, PolicyArn=policy['PolicyArn'])
    print('Deleting all inline policies on the user')
    for policy in get_all_inline_user_policies(name):
        client.delete_user_policy(UserName=name, PolicyName=policy)
    print('Detaching access keys on the user')
    for key in get_all_access_keys_for_user(name):
        client.delete_access_key(UserName=name, AccessKeyId=key['AccessKeyId'])
    print('Deleting user ssh keys')
    for key in get_all_ssh_keys_for_user(name):
        client.delete_ssh_public_key(UserName=name, SSHPublicKeyId=key['SSHPublicKeyId'])
    print('Deleting user mfa devices')
    for device in get_all_mfa_devices_for_user(name):
        client.deactivate_mfa_device(UserName=name, SerialNumber=device['SerialNumber'])
        client.delete_virtual_mfa_device(SerialNumber=device['SerialNumber'])
    print('Deleting user login profile')
    try:
        # Check to see if the user has a login profile before attempting a delete
        client.get_login_profile(UserName=name)
        client.delete_login_profile(UserName=name)
    except client.exceptions.NoSuchEntityException:
        print('User has no login profile. Ignoring error.')

    print('Deleting user')
    client.delete_user(UserName=name)


def get_all_groups_for_user(username):
    """Get all the groups attached to the user"""
    return get_all_objects(
        boto3.client('iam').list_groups_for_user,
        'Groups',
        extra_kwargs={
            'UserName': username,
        },
    )


def get_all_user_policies(username):
    """Get all the policies attached to the user"""
    return get_all_objects(
        boto3.client('iam').list_attached_user_policies,
        'AttachedPolicies',
        extra_kwargs={
            'UserName': username,
        },
    )


def get_all_inline_user_policies(username):
    """Get all the policies declared inline on the user"""
    return get_all_objects(
        boto3.client('iam').list_user_policies,
        'PolicyNames',
        extra_kwargs={
            'UserName': username,
        },
    )


def get_all_access_keys_for_user(username):
    """Get all the access keys associated to the user"""
    return get_all_objects(
        boto3.client('iam').list_access_keys,
        'AccessKeyMetadata',
        extra_kwargs={
            'UserName': username,
        },
    )


def get_all_ssh_keys_for_user(username):
    """Get all the associated ssh keys on the user"""
    return get_all_objects(
        boto3.client('iam').list_ssh_public_keys,
        'SSHPublicKeys',
        extra_kwargs={
            'UserName': username,
        },
    )


def get_all_mfa_devices_for_user(username):
    """Get all the associated mfa devices on the user"""
    return get_all_objects(
        boto3.client('iam').list_mfa_devices,
        'MFADevices',
        extra_kwargs={
            'UserName': username,
        },
    )


# ---------------------------------------------------------------------------------------------------------------------
# HELPER FUNCTIONS TO PROCESS IAM GROUPS
# ---------------------------------------------------------------------------------------------------------------------

def get_all_groups():
    """Get all IAM Groups in the account, accounting for pagination in the API"""
    return get_all_objects(boto3.client('iam').list_groups, 'Groups')


def delete_group(group):
    """Delete the given IAM Group"""
    client = boto3.client('iam')
    name = group['GroupName']
    create_date = group['CreateDate']
    print('Deleting group {} created on {}'.format(name, create_date.isoformat()))
    print('Detaching all policies on the group')
    for policy in get_all_group_policies(name):
        client.detach_group_policy(GroupName=name, PolicyArn=policy['PolicyArn'])
    print('Deleting all inline policies on the group')
    for policy in get_all_inline_group_policies(name):
        client.delete_group_policy(GroupName=name, PolicyName=policy)
    print('Deleting group')
    client.delete_group(GroupName=name)


def get_all_group_policies(group_name):
    """Get all the policies attached to the group"""
    return get_all_objects(
        boto3.client('iam').list_attached_group_policies,
        'AttachedPolicies',
        extra_kwargs={
            'GroupName': group_name,
        },
    )


def get_all_inline_group_policies(group_name):
    """Get all the policies declared inline on the group"""
    return get_all_objects(
        boto3.client('iam').list_group_policies,
        'PolicyNames',
        extra_kwargs={
            'GroupName': group_name,
        },
    )


# ---------------------------------------------------------------------------------------------------------------------
# HELPER FUNCTIONS TO PROCESS IAM ROLES
# ---------------------------------------------------------------------------------------------------------------------

def get_all_roles():
    """Get all IAM Roles in the account, accounting for pagination in the API"""
    return get_all_objects(boto3.client('iam').list_roles, 'Roles')


def delete_role(role):
    """Delete the given IAM Role"""
    client = boto3.client('iam')
    name = role['RoleName']
    create_date = role['CreateDate']
    print('Deleting role {} created on {}'.format(name, create_date.isoformat()))
    print('Detaching all policies on the role')
    for policy in get_all_role_policies(name):
        client.detach_role_policy(RoleName=name, PolicyArn=policy['PolicyArn'])
    print('Deleting all inline policies on the role')
    for policy in get_all_inline_role_policies(name):
        client.delete_role_policy(RoleName=name, PolicyName=policy)
    print('Removing instance profile from role')
    for profile in get_all_associated_instance_profiles_on_role(name):
        client.remove_role_from_instance_profile(InstanceProfileName=profile['InstanceProfileName'], RoleName=name)
    print('Deleting role')
    client.delete_role(RoleName=name)


def get_all_role_policies(role_name):
    """Get all the policies attached to the role"""
    return get_all_objects(
        boto3.client('iam').list_attached_role_policies,
        'AttachedPolicies',
        extra_kwargs={
            'RoleName': role_name,
        }
    )


def get_all_inline_role_policies(role_name):
    """Get all the policies declared inline on the role"""
    return get_all_objects(
        boto3.client('iam').list_role_policies,
        'PolicyNames',
        extra_kwargs={
            'RoleName': role_name,
        }
    )


# ---------------------------------------------------------------------------------------------------------------------
# HELPER FUNCTIONS TO PROCESS IAM INSTANCE PROFILES
# ---------------------------------------------------------------------------------------------------------------------

def get_all_associated_instance_profiles_on_role(role_name):
    """Get all the associated Instance Profiles on the role"""
    return get_all_objects(
        boto3.client('iam').list_instance_profiles_for_role,
        'InstanceProfiles',
        extra_kwargs={
            'RoleName': role_name,
        }
    )


def get_all_profiles():
    """Get all IAM Instance Profiles in the account, accounting for pagination in the API"""
    return get_all_objects(boto3.client('iam').list_instance_profiles, 'InstanceProfiles')


def delete_profile(profile):
    """Delete the given IAM Instance Profile"""
    client = boto3.client('iam')
    name = profile['InstanceProfileName']
    create_date = profile['CreateDate']
    print('Deleting profile {} created on {}'.format(name, create_date.isoformat()))
    print('Removing roles from instance profile')
    for role in profile['Roles']:
        client.remove_role_from_instance_profile(InstanceProfileName=name, RoleName=role['RoleName'])
    print('Deleting instance profile')
    client.delete_instance_profile(InstanceProfileName=name)


# ---------------------------------------------------------------------------------------------------------------------
# GENERAL HELPER FUNCTIONS
# ---------------------------------------------------------------------------------------------------------------------

def get_all_objects(getter, response_key, extra_args=None, extra_kwargs=None):
    """
    Given a function that conforms to the boto api of a paginated list function, repeatedly call it until there are no
    more pages to retrieve.

    Args:
        getter : A function that can be called to retrieve the objects. Should take in a kwarg for the pagination
                 marker, and return a list for its response.
        response_key : The key to use to extract the list of objects from the list response.
        extra_args : The list of function args to pass to the getter.
        extra_kwargs : The list of function kwargs to pass to the getter.

    Returns:
        The list of objects from the getter.
    """
    # We use None instead of directly setting [] or {} in function params as the defaults, so that we create a new
    # list/dict each time. Otherwise, the list or dict is globally scoped, so modifications to the initial args will
    # persist across calls.
    if extra_args is None:
        extra_args = []
    if extra_kwargs is None:
        extra_kwargs = {}

    response = getter(*extra_args, **extra_kwargs)
    objs = response[response_key]
    while 'Marker' in response:
        extra_kwargs['Marker'] = response['Marker']
        response = getter(*extra_args, **extra_kwargs)
        objs.extend(response[response_key])
    return objs


def is_test_role_or_instance_profile(name):
    """Whether or not the given name matches a heuristic list of known test IAM names"""
    regex_list = [
        r'^cloud-nuke-test-.+',
        r'^cloud-nuke-Test.+',
        r'^influxdb-data-[a-zA-Z0-9]{6}\d+.$',
        r'^influxdb-meta-[a-zA-Z0-9]{6}\d+.$',
        r'^influxdb-[a-zA-Z0-9]{6}\d+.$',
        r'^kapacitor-server-[a-zA-Z0-9]{6}\d+.$',
        r'^vault-test-[a-zA-Z0-9]{6}.+',
        r'^consul-test-[a-zA-Z0-9]{6}.+',
        r'^kibana-[a-zA-Z0-9]{6}.+',
        r'^server-group.+',
        r'^[Tt]est-cluster[a-zA-Z0-9]{6}$',
        r'^[a-zA-Z0-9]{6}-cluster$',
        r'^[a-zA-Z0-9]{6}-ecs-cluster$',
        r'(app|core)-workers-eks-cluster-[a-zA-Z0-9]{6}-worker$',
        r'eks-cluster-[a-zA-Z0-9]{6}-(cluster|worker)$',
    ]
    return any(re.match(regex, name) for regex in regex_list)


def is_test_user_or_group(name):
    """Whether or not the given name matches a heuristic list of known test IAM Usernames and Group names"""
    regex_list = [
        r'^cross-account-iam-roles-test-[a-zA-Z0-9]{6}-[a-zA-Z0-9]{6}$',
        r'^test-user-non-sudo-[a-zA-Z0-9]{6}$',
        r'^test-user-sudo-[a-zA-Z0-9]{6}$',
        r'^test-group-non-sudo-[a-zA-Z0-9]{6}$',
        r'^test-group-sudo-[a-zA-Z0-9]{6}$',
        r'^TestIamGroupUseExistingIamRoles-[a-zA-Z0-9]{6}$',
        r'^TestIamPolicyIamUserSelfMgmt-[a-zA-Z0-9]{6}$',
        r'^tst-openvpn-host-(\d+|[a-zA-Z0-9]{6})-(Admins|Users)$',
    ]
    return any(re.match(regex, name) for regex in regex_list)


def created_before_yesterday(t):
    """Whether or not the given time is before yesterday (24h ago)"""
    unow = dt.datetime.utcnow().replace(tzinfo=pytz.utc)
    return t < unow - dt.timedelta(hours=24)


def want_profile(profile):
    """Which IAM Instance Profiles we want to delete

    Args:
        profile (dict) : Dictionary representation of an IAM Instance Profile as returned by boto3.

    Returns:
        Boolean indicating whether or not we want to delete the given IAM Instance Profile.
    """
    name = profile['InstanceProfileName']
    create_date = profile['CreateDate']
    return is_test_role_or_instance_profile(name) and created_before_yesterday(create_date)


def want_role(role):
    """Which IAM Roles we want to delete

    Args:
        role (dict) : Dictionary representation of an IAM Role as returned by boto3.

    Returns:
        Boolean indicating whether or not we want to delete the given IAM Role.
    """
    name = role['RoleName']
    create_date = role['CreateDate']
    return is_test_role_or_instance_profile(name) and created_before_yesterday(create_date)


def want_user(role):
    """Which IAM Users we want to delete

    Args:
        role (dict) : Dictionary representation of an IAM Role as returned by boto3.

    Returns:
        Boolean indicating whether or not we want to delete the given IAM User.
    """
    name = role['UserName']
    create_date = role['CreateDate']
    return is_test_user_or_group(name) and created_before_yesterday(create_date)


def want_group(role):
    """Which IAM Groups we want to delete

    Args:
        role (dict) : Dictionary representation of an IAM Role as returned by boto3.

    Returns:
        Boolean indicating whether or not we want to delete the given IAM User.
    """
    name = role['GroupName']
    create_date = role['CreateDate']
    return is_test_user_or_group(name) and created_before_yesterday(create_date)


# ---------------------------------------------------------------------------------------------------------------------
# RUN FUNCTIONS: MAIN ENTRYPOINTS FOR EACH LOGIC
# ---------------------------------------------------------------------------------------------------------------------

def run_users(dry=True):
    """Run the garbage collection routine for Users

    This will:
    - Find all the users in the account
    - Filter the test users based on some heuristics on the name
    - List out all the users it found in a table
    - If in "wet" mode, ask for confirmation from the operator to proceed with the deletion
    """
    users = [user for user in get_all_users() if want_user(user)]
    print('Found {} users'.format(len(users)))
    if len(users) == 0:
        return

    print('Last user created {}'.format(max(user['CreateDate'] for user in users).isoformat()))
    users.sort(key=lambda p: p['CreateDate'], reverse=True)
    print(
        tabulate(
            [(user['UserName'], user['CreateDate'].isoformat()) for user in users],
            headers=('Name', 'Created'),
        )
    )
    print()

    if dry:
        return

    input('Will delete {} users. [Ctrl+C] to cancel, or [ENTER] to proceed.'.format(len(users)))
    bar = Bar('Deleting', max=len(users))
    for user in bar.iter(users):
        delete_user(user)


def run_groups(dry=True):
    """Run the garbage collection routine for Users

    This will:
    - Find all the groups in the account
    - Filter the test groups based on some heuristics on the name
    - List out all the groups it found in a table
    - If in "wet" mode, ask for confirmation from the operator to proceed with the deletion
    """
    groups = [group for group in get_all_groups() if want_group(group)]
    print('Found {} groups'.format(len(groups)))
    if len(groups) == 0:
        return

    print('Last group created {}'.format(max(group['CreateDate'] for group in groups).isoformat()))
    groups.sort(key=lambda p: p['CreateDate'], reverse=True)
    print(
        tabulate(
            [(group['GroupName'], group['CreateDate'].isoformat()) for group in groups],
            headers=('Name', 'Created'),
        )
    )
    print()

    if dry:
        return

    input('Will delete {} groups. [Ctrl+C] to cancel, or [ENTER] to proceed.'.format(len(groups)))
    bar = Bar('Deleting', max=len(groups))
    for group in bar.iter(groups):
        delete_group(group)


def run_roles(dry=True):
    """Run the garbage collection routine for Roles

    This will:
    - Find all the roles in the account
    - Filter the test roles based on some heuristics on the name
    - List out all the roles it found in a table
    - If in "wet" mode, ask for confirmation from the operator to proceed with the deletion
    """
    roles = [role for role in get_all_roles() if want_role(role)]
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
    """Run the garbage collection routine for Instance Profiles.

    This will:
    - Find all the profiles in the account
    - Filter the test profiles based on some heuristics on the name
    - List out all the profiles it found in a table
    - If in "wet" mode, ask for confirmation from the operator to proceed with the deletion
    """
    profiles = [profile for profile in get_all_profiles() if want_profile(profile)]
    print('Found {} profiles'.format(len(profiles)))
    if len(profiles) == 0:
        return

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


def parse_args():
    """Parse command line args"""
    parser = argparse.ArgumentParser(
        description='This script garbage collects test IAM roles and instance profiles.')

    parser.add_argument('-r', '--run', action='store_true', help='Run the deletion routine.')

    args = parser.parse_args()
    return args


def main():
    args = parse_args()
    dry = not args.run

    run_users(dry)
    run_groups(dry)
    run_profiles(dry)
    run_roles(dry)


if __name__ == '__main__':
    main()
