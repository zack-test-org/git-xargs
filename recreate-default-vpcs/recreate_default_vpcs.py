import boto3


def get_all_regions():
    client = boto3.client('ec2', region_name='us-east-1')
    return [region['RegionName'] for region in client.describe_regions()['Regions']]


def has_default_vpc(region):
    client = boto3.client('ec2', region_name=region)
    filters = [{'Name': 'isDefault', 'Values': ['true']}]
    vpcs = client.describe_vpcs(Filters=filters)['Vpcs']
    return len(vpcs) > 0


def create_default_vpc(region):
    client = boto3.client('ec2', region_name=region)
    return client.create_default_vpc()['Vpc']


def create_default_vpc_if_not_exist(region):
    if has_default_vpc(region):
        print('Region {} has default VPC. Skipping create.'.format(region))
        return

    print('Region {} does not have default VPC. Creating.'.format(region))
    vpc = create_default_vpc(region)
    print('Created VPC {}'.format(vpc['VpcId']))


def main():
    print('Retrieving all enabled regions')
    regions = get_all_regions()

    print('Found {} regions'.format(len(regions)))
    print('Ensuring default VPC exists in each region')
    for region in regions:
        create_default_vpc_if_not_exist(region)


if __name__ == '__main__':
    main()
