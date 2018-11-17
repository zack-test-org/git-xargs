import boto3
import collections
import json
import re
import six

from . import global_vars


def get_region_description(region):
    """
    Given a region code (e.g us-east-1) return region description (e.g North Virginia).
    """
    for partition in global_vars.aws_endpoints["partitions"]:
        if region in partition["regions"]:
            return partition["regions"][region]["description"]
    return None


def get_available_instance_types(region):
    """
    Given a region code (e.g us-east-1), return all the instance types available in the region.
    """
    product_pager = global_vars.pricing_client.get_paginator("get_products")
    ec2_iterator = product_pager.paginate(
        ServiceCode="AmazonEC2",
        Filters=[
            # Pricing API filters by region description, not region code
            {"Type": "TERM_MATCH", "Field": "location", "Value": get_region_description(region)},
            # Do not want dedicated only instance types
            {'Type': 'TERM_MATCH', 'Field': 'tenancy', 'Value': 'Shared'},
            # Do not want anything that requires special licensing
            {'Type': 'TERM_MATCH', 'Field': 'licenseModel', 'Value': 'No License required'},
        ]
    )
    instance_types = []
    for product_item in ec2_iterator:
        for offer_string in product_item.get("PriceList"):
            offer = json.loads(offer_string)
            product = offer.get("product")

            # Check if it's an instance
            if product.get("productFamily") != "Compute Instance":
                continue

            product_attributes = product.get("attributes")
            instance_type = product_attributes.get("instanceType")
            instance_types.append(instance_type)
    return instance_types


def get_available_rds_instance_types(region):
    """
    Given a region code (e.g us-east-1), return all the RDS instance types available in the region.

    Returns as dictionary mapping database engines to available instance classes.
    """
    product_pager = global_vars.pricing_client.get_paginator("get_products")
    rds_iterator = product_pager.paginate(
        ServiceCode="AmazonRDS",
        Filters=[
            # Pricing API filters by region description, not region code
            {"Type": "TERM_MATCH", "Field": "location", "Value": get_region_description(region)},
        ]
    )
    instance_types = collections.defaultdict(list)
    for product_item in rds_iterator:
        for offer_string in product_item.get("PriceList"):
            offer = json.loads(offer_string)
            product = offer.get("product")

            # Check if it's an instance
            if product.get("productFamily") != "Database Instance":
                continue

            product_attributes = product.get("attributes")
            db_engine = product_attributes.get("databaseEngine")
            db_engine_edition = product_attributes.get("databaseEdition")
            db_code = get_rds_engine_code(db_engine, db_engine_edition)
            if db_code is None:
                continue

            instance_type = product_attributes.get("instanceType")
            instance_types[db_code].append(instance_type)
    return instance_types


def get_available_cache_instance_types(region):
    """
    Given a region code (e.g us-east-1), return all the ElastiCache instance types available in the region.

    Returns as dictionary mapping cache engines to available instance classes.
    """
    product_pager = global_vars.pricing_client.get_paginator("get_products")
    ec_iterator = product_pager.paginate(
        ServiceCode="AmazonElastiCache",
        Filters=[
            # Pricing API filters by region description, not region code
            {"Type": "TERM_MATCH", "Field": "location", "Value": get_region_description(region)},
        ]
    )
    instance_types = collections.defaultdict(list)
    for product_item in ec_iterator:
        for offer_string in product_item.get("PriceList"):
            offer = json.loads(offer_string)
            product = offer.get("product")

            # Check if it's an instance
            if product.get("productFamily") != "Cache Instance":
                continue

            product_attributes = product.get("attributes")
            cache_engine = product_attributes.get("cacheEngine")
            if cache_engine not in ["Redis", "Memcached"]:
                continue

            instance_type = product_attributes.get("instanceType")
            instance_types[cache_engine].append(instance_type)
    return instance_types


def find_domains_with_no_certs(environment_credentials, region, domains_to_check):
    """
    Using the environment credentials, find all the domains that don't have ACM certificates set in the region.
    """
    domains_with_no_certs = []
    for account, domain in six.iteritems(domains_to_check):
        credentials = environment_credentials[account]
        acm_domains = get_acm_certificate_domains(credentials, region)
        domain_as_certificate_domain_name = "*.{}".format(domain)
        if domain_as_certificate_domain_name not in acm_domains:
            domains_with_no_certs.append((account, domain))
    return domains_with_no_certs


def get_hosted_zone_names(credentials):
    """
    Look up all the hosted zones in an account and return the domain names.
    """
    route53 = get_client_by_credentials(credentials, "route53")
    response = route53.list_hosted_zones()
    # boto3 will raise an exception on bad response, so if we make it this far, can assume the key exists.
    zones = response["HostedZones"]
    zone_names = [zone["Name"] for zone in zones]
    return zone_names


def get_acm_certificate_domains(credentials, region):
    """
    Look up all the ACM certificates in an account for the given region and return the domain names.
    """
    acm = get_client_by_credentials(credentials, "acm", region)
    response = acm.list_certificates()
    # boto3 will raise an exception on bad response, so if we make it this far, can assume the key exists.
    acm_certificates = response["CertificateSummaryList"]
    certificate_domain_names = [acm_cert["DomainName"] for acm_cert in acm_certificates]
    return certificate_domain_names


def get_client_by_credentials(credentials, aws_app, region=None):
    """
    Given a set of credentials specified in a dictionary, return the authenticated sts client.
    """
    client = boto3.client(
        aws_app,
        region_name=region,
        aws_access_key_id=credentials["AccessKeyId"],
        aws_secret_access_key=credentials["SecretAccessKey"],
        aws_session_token=credentials["SessionToken"],
    )
    return client


def get_environment_credentials(parsed_yamls):
    """
    Given a dictionary of parsed yamls for the reference architecture, return a map of temporary credentials that can be
    used to access the environments. Each credential can be used to assume the GruntworkAccountAccessRole in each
    account.
    """
    accounts = None
    for parsed_yaml in parsed_yamls.values():
        if parsed_yaml is not None and "Accounts" in parsed_yaml:
            accounts = parsed_yaml["Accounts"]
            break

    credentials = {}
    for session_name, customer_account_id in six.iteritems(accounts):
        credentials[session_name] = obtain_credentials_to_assume_gruntwork_role(
            customer_account_id,
            session_name)
    return credentials


def obtain_credentials_to_assume_gruntwork_role(account_id, session_name):
    """
    Returns credentials to assume the GruntworkAccountAccessRole in the customer's account provided by `account_id`.

    Returns:
        Credentials dictionary with the following schema:
            {
                "AccessKeyId": "AWS Access Key ID to use to assume the requested role.",
                "SecretAccessKey": "AWS Secret Access Key to use to assume the requested role.",
                "SessionToken": "STS Session token to use to assume the requested role.",
            }
    """
    role_arn = "arn:aws:iam::{}:role/GruntworkAccountAccessRole".format(account_id)
    global_vars.logger.debug(
        "Requesting creds for role {} in account {} ({})".format(role_arn, account_id, session_name))
    assume_role_response = global_vars.sts_client.assume_role(
        RoleArn=role_arn,
        RoleSessionName=session_name)
    # boto3 will raise an exception on bad response, so if we make it this far, can assume the key exists.
    return assume_role_response["Credentials"]


def get_account_info():
    """
    Extract the account id and username based on currently configured AWS credentials.

    Returns:
        Tuple of (AWS Account ID, Username)
    """
    caller_identity = global_vars.sts_client.get_caller_identity()
    name_re = r".+/([^/]+)$"
    username = re.match(name_re, caller_identity["Arn"]).group(1)
    return caller_identity["Account"], username


def get_rds_engine_code(engine_description, engine_edition):
    """
    Given engine description (e.g SQL Server) and engine edition (e.g Enterprise), return engine code (e.g
    sqlserver-ee).
    """
    if engine_description == "Aurora MySQL":
        return "aurora-mysql"
    elif engine_description == "Aurora PostgreSQL":
        return "aurora-postgresql"
    elif engine_description == "MariaDB":
        return "mariadb"
    elif engine_description == "MySQL":
        return "mysql"
    elif engine_description == "PostgreSQL":
        return "postgres"
    elif engine_description == "SQL Server":
        if engine_edition == "Enterprise":
            return "sqlserver-ee"
        elif engine_edition == "Express":
            return "sqlserver-ex"
        elif engine_edition == "Standard":
            return "sqlserver-se"
        elif engine_edition == "Web":
            return "sqlserver-web"
        else:
            global_vars.logger.warn(
                "Unknown engine edition {} from pricing API for SQL server".format(engine_edition))
            return None
    elif engine_description == "Oracle":
        if engine_edition == "Enterprise":
            return "oracle-ee"
        elif engine_edition == "Standard":
            return "oracle-se"
        elif engine_edition == "Standard One":
            return "oracle-se1"
        elif engine_edition == "Standard Two":
            return "oracle-se1"
        else:
            global_vars.logger.warn(
                "Unknown engine edition {} from pricing API for Oracle DB".format(engine_edition))
            return None
    else:
        global_vars.logger.warn("Unknown engine {} from pricing API".format(engine_description))
        return None
