import re
import boto3
from tabulate import tabulate
from progress.bar import Bar
from collections import namedtuple

ECSCluster = namedtuple('ECSCluster', ['arn', 'region'])


def run():
    """
    Find all ECS clusters across all enabled regions that are considered test clusters, and delete them.

    A test cluster is:
    - Any cluster that matches one of the regexexs in is_test_ecs_cluster.
    - Has no ECS service deployed to it.
    """
    ec2 = boto3.client('ec2', 'us-east-1')
    enabled_regions = ec2.describe_regions()

    ecs_clusters = []
    for region in enabled_regions['Regions']:
        region_name = region['RegionName']
        print(f'Looking up ECS clusters in region {region_name}')
        ecs_clusters_in_region = [
            ECSCluster(carn, region_name) for carn in get_all_ecs_clusters(region_name)
            if want_ecs_cluster(region_name, carn)
        ]
        print(f'Found {len(ecs_clusters_in_region)} clusters in region {region_name}')
        ecs_clusters.extend(ecs_clusters_in_region)
    if len(ecs_clusters) == 0:
        return

    print(
        tabulate(
            [(cluster.region, cluster.arn) for cluster in ecs_clusters],
            headers=('Region', 'ARN'),
        )
    )
    print()

    input(f'Will delete {len(ecs_clusters)} ECS clusters. [Ctrl+C] to cancel, or [ENTER] to proceed.')
    bar = Bar('Deleting', max=len(ecs_clusters))
    for cluster in bar.iter(ecs_clusters):
        delete_ecs_cluster(cluster.region, cluster.arn)


def get_all_ecs_clusters(region):
    """ Get all ECS clusters in the configured region, accounting for pagination. """
    ecs_client = boto3.client('ecs', region_name=region)
    response = ecs_client.list_clusters()
    clusters = response['clusterArns']
    while 'nextToken' in response and response['nextToken']:
        response = ecs_client.list_clusters(nextToken=response['nextToken'])
        clusters.extend(response['clusterArns'])
    return clusters


def delete_ecs_cluster(region, cluster_arn):
    """ Delete the provided ECS cluster. """
    ecs_client = boto3.client('ecs', region_name=region)
    ecs_client.delete_cluster(cluster=cluster_arn)


def is_test_ecs_cluster(cluster_arn):
    """
    Checks if the provided cluster in the given region is a test ECS cluster. We use a name based heuristic to decide if
    the given cluster is a test cluster.
    """
    regex_list = [
        r'^cloud-nuke-test-[a-zA-Z0-9]{6}-cluster$',
        r'^[a-zA-Z0-9]{6}-cluster$',
        r'^test-cluster[a-zA-Z0-9]{6}$',
        r'^[a-zA-Z0-9]{6}-ecs-cluster$',
        r'^Test-cluster[a-zA-Z0-9]{6}$',
    ]
    cluster_name = cluster_arn.split('/')[1]
    return any(re.match(regex, cluster_name) for regex in regex_list)


def cluster_is_empty(region, cluster_arn):
    """
    Checks if the cluster is empty. A cluster is empty if it has no ECS service deployed to it.
    """
    ecs_client = boto3.client('ecs', region_name=region)
    services = ecs_client.list_services(cluster=cluster_arn)
    return len(services['serviceArns']) == 0


def want_ecs_cluster(region, cluster_arn):
    """
    We want to delete clusters that are:
    - A test ECS Cluster.
    - Has no ECS service deployed to it.
    """
    return is_test_ecs_cluster(cluster_arn) and cluster_is_empty(region, cluster_arn)


if __name__ == '__main__':
    run()
