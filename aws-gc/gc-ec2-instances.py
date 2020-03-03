import re
import boto3
from tabulate import tabulate
from collections import namedtuple

Instance = namedtuple('Instance', ['region', 'info'])


def run():
    """
    Find all EC2 instances that have termination protection on with test names and disable termination protection on the
    instances.
    """
    ec2 = boto3.client('ec2', 'us-east-1')
    enabled_regions = ec2.describe_regions()

    instances = []
    for region in enabled_regions['Regions']:
        region_name = region['RegionName']
        print(f'Looking up instances in region {region_name}')
        instances_in_region = [
            Instance(region_name, inst) for inst in get_all_running_ec2_instances(region_name)
            if want_ec2_instance(region_name, inst)
        ]
        print(f'Found {len(instances_in_region)} instances in region {region_name}')
        instances.extend(instances_in_region)
    if len(instances) == 0:
        return

    print('Last instance created {}'.format(max(inst.info['LaunchTime'] for inst in instances).isoformat()))
    print(
        tabulate(
            [
                (
                    inst.region,
                    inst.info['LaunchTime'].isoformat(),
                    inst.info['InstanceId'],
                    get_instance_tag_by_key(inst.info, 'Name'),
                )
                for inst in instances
            ],
            headers=('Region', 'Launch Time', 'Id', 'Name'),
        )
    )
    print()

    input(f'Will remove termination protection on {len(instances)} instances. [Ctrl+C] to cancel, or [ENTER] to proceed.')
    for instance in instances:
        remove_termination_protection(instance.region, instance.info)


def remove_termination_protection(region, instance):
    """ Remove termination protection from the given instance in the provided region. """
    ec2_client = boto3.client('ec2', region_name=region)
    ec2_client.modify_instance_attribute(
        InstanceId=instance['InstanceId'],
        Attribute='disableApiTermination',
        DisableApiTermination={'Value': False},
    )


def get_all_running_ec2_instances(region):
    """ Get all EC2 instances that are running in the provided region. """
    ec2_client = boto3.client('ec2', region_name=region)
    kwargs = {
        'Filters': [
            {
                'Name': 'instance-state-name',
                'Values': ['running'],
            },
        ]
    }
    response = ec2_client.describe_instances(**kwargs)
    instances = sum([res['Instances'] for res in response['Reservations']], [])
    while 'NextToken' in response and response['NextToken']:
        kwargs['NextToken'] = response['NextToken']
        response = ec2_client.describe_instances(**kwargs)
        instances.extend(sum([res['Instances'] for res in response['Reservations']], []))
    return instances


def is_termination_protected(region, instance):
    """ Returns true if the given instance in the provided region has termination protection enabled. """
    ec2_client = boto3.client('ec2', region_name=region)
    response = ec2_client.describe_instance_attribute(
        Attribute='disableApiTermination', InstanceId=instance['InstanceId']
    )
    return response['DisableApiTermination']


def is_test_instance(instance):
    """
    Returns true if the provided instance is a test instance, as determined by the name. Any instance with no name is
    also considered a test instance.
    """
    regex_list = [
        r'^cloud-nuke-test-[a-zA-Z0-9]{6}$',
        r'^$',
    ]
    return any(
        get_instance_tag_by_key(instance, 'Name') is None or re.match(regex, get_instance_tag_by_key(instance, 'Name'))
        for regex in regex_list
    )


def get_instance_tag_by_key(instance, key):
    """ Get the value of the tag on the instance with the given key. """
    for tag in instance.get('Tags', []):
        if tag['Key'] == key:
            return tag['Value']
    return None


def want_ec2_instance(region, instance):
    """
    We want to disable termination protection on instances that are a test instance and that have termination protection
    enabled on the instance.
    """
    return is_test_instance(instance) and is_termination_protected(region, instance)


if __name__ == '__main__':
    run()
