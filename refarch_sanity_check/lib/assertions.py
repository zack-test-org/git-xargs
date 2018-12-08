import click
import six

from . import aws, global_vars, utils, parser


def assert_acme_variables_are_deleted(parsed_yamls):
    """
    Checks that the following variables are not defined:
    - IsLegacyCustomerOrTest
    - stage-and-prod-sample-frontend-app.ServiceName
    - stage-and-prod-sample-backend-app.ServiceName
    - GitBranch
    """
    _CHECK_VARS = [
        "IsLegacyCustomerOrTest",
        "stage-and-prod-sample-frontend-app.ServiceName",
        "stage-and-prod-sample-backend-app.ServiceName",
        "GitBranch",
    ]
    all_keys = utils.all_nested_keys(parsed_yamls)
    acme_vars_found = [
        var for var in _CHECK_VARS if var in all_keys
    ]
    if acme_vars_found:
        error_string = "\n".join(
            "\t{}".format(var)
            for var in acme_vars_found
        )
        raise click.ClickException("Found Acme only vars in config:\n{}".format(error_string))


def assert_ecr_repositories_configured(parsed_yamls):
    """
    Checks that the ECR repositories for the sample apps are pointing to the right account and region.
    """
    if len(parsed_yamls) == 1:
        # Single account doesn't have this risk
        return

    root_vars = parsed_yamls["root-vars.yml"]
    accounts = root_vars["Accounts"]
    if "shared-services" not in accounts:
        global_vars.logger.warning(
            "Skipping ECR repository check because there is no shared-services account configured.")
        return
    shared_services_account_id = accounts["shared-services"]
    region = parser.get_configured_region(parsed_yamls)

    expected_repository_domain_name = "{}.dkr.ecr.{}.amazonaws.com".format(shared_services_account_id, region)
    IMAGE_NAME_KEYS = ["FrontendDockerImageName", "BackendDockerImageName"]
    misconfigured_ecr_repositories = []
    for key in IMAGE_NAME_KEYS:
        if key in root_vars and not root_vars[key].startswith(expected_repository_domain_name):
            misconfigured_ecr_repositories.append(key)

    if misconfigured_ecr_repositories:
        for key in misconfigured_ecr_repositories:
            global_vars.logger.error(
                "ECR repository config {} is not pointing to the right account and region (expected domain {})"
                .format(key, expected_repository_domain_name))
        raise click.ClickException("Failed ECR repository check")


def assert_route_53_domains_exist(environment_credentials, parsed_yamls):
    """
    Checks that the proper route 53 DNS entry exists for the specified domain names in the vars file.
    """
    domains_to_check = parser.get_configured_domain_names(parsed_yamls)
    nonexistant_domains = []
    for account, domain in six.iteritems(domains_to_check):
        credentials = environment_credentials[account]
        existing_zone_names = aws.get_hosted_zone_names(credentials)
        domain_as_zone_name = "{}.".format(domain)
        if domain_as_zone_name not in existing_zone_names:
            nonexistant_domains.append((account, domain))

    if nonexistant_domains:
        for account, domain in nonexistant_domains:
            global_vars.logger.error("Could not find domain in account {} (domain {})".format(account, domain))
        click.ClickException("Failed domain check")


def assert_acm_certificates_exist(environment_credentials, parsed_yamls):
    """
    Checks that ACM certificates exist in each account for each domain.
    """
    region = parser.get_configured_region(parsed_yamls)
    domains_to_check = parser.get_configured_domain_names(parsed_yamls)
    domains_with_no_certs = aws.find_domains_with_no_certs(environment_credentials, region, domains_to_check)

    # If static websites are configured, then check in us-east-1 too since cloudfront can only use us-east-1 for ACM
    # certs.
    if parser.is_static_websites_configured(parsed_yamls):
        # Filter out shared services account
        static_domains_to_check = {
            account: domain for account, domain in six.iteritems(domains_to_check)
            if account != "shared-services"
        }
        domains_with_no_certs += aws.find_domains_with_no_certs(
            environment_credentials, "us-east-1", static_domains_to_check)

    if domains_with_no_certs:
        for account, domain in domains_with_no_certs:
            global_vars.logger.error(
                "Could not find ACM certificate in account {} for domain {}".format(account, domain))
        raise click.ClickException("Failed certificate check")


def assert_instance_types_available(parsed_yamls):
    """
    Checks to make sure instance types configured are available in the selected region.
    """
    instance_types_to_check = parser.get_configured_instance_types(parsed_yamls)

    global_vars.logger.debug("Retrieving instance availability information. This takes some time...")
    region = parser.get_configured_region(parsed_yamls)
    available_instance_types = aws.get_available_instance_types(region)
    global_vars.logger.debug("Done retrieving instance availability information.")

    unavailable_instance_types = []
    for instance_type in instance_types_to_check:
        if instance_type not in available_instance_types:
            unavailable_instance_types.append(instance_type)

    if unavailable_instance_types:
        for instance_type in unavailable_instance_types:
            global_vars.logger.error("Instance type {} is not available in region {}".format(instance_type, region))
        raise click.ClickException("Failed instance type check")


def assert_rds_instance_types_available(parsed_yamls):
    """
    Checks to make sure RDS instance types configured are available in the selected region.
    """
    instance_types_to_check = parser.get_configured_rds_instance_types(parsed_yamls)

    global_vars.logger.debug("Retrieving RDS instance availability information. This takes some time...")
    region = parser.get_configured_region(parsed_yamls)
    available_rds_instance_types_by_engines = aws.get_available_rds_instance_types(region)
    global_vars.logger.debug("Done retrieving RDS instance availability information.")

    rds_engine = parser.get_configured_database_engine(parsed_yamls)
    if rds_engine is None or rds_engine not in available_rds_instance_types_by_engines:
        global_vars.logger.error("DB engine {} is unrecognized in region {}".format(rds_engine, region))
        raise click.ClickException("DB engine is misconfigured")

    available_rds_instance_types = available_rds_instance_types_by_engines[rds_engine]
    unavailable_rds_instance_types = []
    for instance_type in instance_types_to_check:
        if instance_type not in available_rds_instance_types:
            unavailable_rds_instance_types.append(instance_type)

    if unavailable_rds_instance_types:
        for instance_type in unavailable_rds_instance_types:
            global_vars.logger.error("RDS Instance type {} is not available in region {}".format(instance_type, region))
        raise click.ClickException("Failed RDS instance type check")


def assert_cache_instance_types_available(parsed_yamls):
    """
    Checks to make sure RDS instance types configured are available in the selected region.
    """
    instance_types_to_check = parser.get_configured_cache_instance_types(parsed_yamls)

    global_vars.logger.debug("Retrieving ElastiCache instance availability information. This takes some time...")
    region = parser.get_configured_region(parsed_yamls)
    available_cache_instance_types_by_engines = aws.get_available_cache_instance_types(region)
    global_vars.logger.debug("Done retrieving ElastiCache instance availability information.")

    cache_engine = parser.get_configured_cache_engine(parsed_yamls)
    if cache_engine is None or cache_engine not in available_cache_instance_types_by_engines:
        global_vars.logger.error("{} is unrecognized in region {}".format(cache_engine, region))
        raise click.ClickException("Cache engine is misconfigured")

    available_cache_instance_types = available_cache_instance_types_by_engines[cache_engine]
    unavailable_cache_instance_types = []
    for instance_type in instance_types_to_check:
        if instance_type not in available_cache_instance_types:
            unavailable_cache_instance_types.append(instance_type)

    if unavailable_cache_instance_types:
        for instance_type in unavailable_cache_instance_types:
            global_vars.logger.error(
                "Cache instance type {} is not available in region {}".format(instance_type, region))
        raise click.ClickException("Failed cache instance type check")


def assert_app_names_different(parsed_yamls):
    """
    Make sure the sample app names are different.
    """
    frontend_app_name, backend_app_name = parser.get_configured_app_names(parsed_yamls)
    if frontend_app_name == backend_app_name:
        raise click.ClickException("The FrontendAppName and the BackendAppName values are the same.")
