import boto3

_KNOWN_EXISTING_REGION = 'us-east-1'


def get_all_regions():
    client = boto3.client('ec2', region_name=_KNOWN_EXISTING_REGION)
    return [region['RegionName'] for region in client.describe_regions()['Regions']]


def has_default_vpc(ec2_client):
    filters = [{'Name': 'isDefault', 'Values': ['true']}]
    vpcs = ec2_client.describe_vpcs(Filters=filters)['Vpcs']
    return len(vpcs) > 0


def create_default_vpc(ec2_client):
    return ec2_client.create_default_vpc()['Vpc']


def create_default_vpc_if_not_exist(ec2_client, region):
    if has_default_vpc(ec2_client):
        print('Region {} has default VPC. Skipping create.'.format(region))
        return

    print('Region {} does not have default VPC. Creating.'.format(region))
    vpc = create_default_vpc(ec2_client)
    print('Created VPC {}'.format(vpc['VpcId']))


def main():
    print('Retrieving all enabled regions')
    regions = get_all_regions()

    print('Found {} regions'.format(len(regions)))
    print('Ensuring default VPC exists in each region')
    for region in regions:
        client = boto3.client('ec2', region_name=region)
        create_default_vpc_if_not_exist(client)


if __name__ == '__main__':
    main()
