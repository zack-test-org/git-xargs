import glob
import os
import re
import six
import yaml

from . import utils


def load_yaml_vars(usage_patterns_path, customer_name):
    """
    Given the path to the usage-patterns repo and the target customer name, generate a dictionary mapping yaml files in
    the directory to the parsed yaml data.

    Returns:
        Parsed YAML dictionaries keyed by the source filename.
    """
    vars_source_path = os.path.join(usage_patterns_path, "stacks", "clients", customer_name)
    yaml_file_names = glob.glob(os.path.join(vars_source_path, "*.yml"))
    parsed_yamls = {}
    for fname in yaml_file_names:
        with open(fname) as f:
            parsed_yamls[os.path.basename(fname)] = yaml.safe_load(f)
    return parsed_yamls


def get_configured_domain_names(parsed_yamls):
    """
    Given the parsed yaml files, extract the domain names configured.
    """
    if len(parsed_yamls) == 1:
        # Single account, so only check the one
        domains_to_check = parsed_yamls["vars.yml"]["DomainNames"]
    else:
        # Multi account, so need to check each domain in each account
        domains_to_check = parsed_yamls["root-vars.yml"]["DomainNames"]
    return domains_to_check


def is_static_websites_configured(parsed_yamls):
    """
    Given the parsed yaml files, determine if static websites are configured.
    """
    for key, value in utils.all_nested_items(parsed_yamls):
        if key == "IncludeStaticAssets":
            return value
    return False


def get_configured_region(parsed_yamls):
    """
    Given the parsed yaml files, extract the region configured.
    """
    if len(parsed_yamls) == 1:
        # Single account
        return parsed_yamls["vars.yml"]["AwsRegion"]
    return parsed_yamls["root-vars.yml"]["AwsRegion"]


def get_configured_instance_types(parsed_yamls):
    """
    Look up all the instance types referenced in the vars.
    """
    instance_type_re = r"^[a-zA-Z][\d]+[a-zA-Z]?\.\d*[a-zA-Z]+$"
    instance_types = set()
    for data in utils.all_nested_values(parsed_yamls):
        if isinstance(data, six.text_type) and re.match(instance_type_re, data):
            instance_types.add(data)
    return set(instance_types)


def get_configured_rds_instance_types(parsed_yamls):
    """
    Look up all the rds instance types referenced in the vars.
    """
    rds_instance_type_re = r"^db.[a-zA-Z][\d]+[a-zA-Z]?\.\d*[a-zA-Z]+$"
    rds_instance_types = set()
    for data in utils.all_nested_values(parsed_yamls):
        if isinstance(data, six.text_type) and re.match(rds_instance_type_re, data):
            rds_instance_types.add(data)
    return set(rds_instance_types)


def get_configured_database_engine(parsed_yamls):
    """
    Look up RDS database engine referenced in the vars.
    """
    for key, value in utils.all_nested_items(parsed_yamls):
        if key == "DatabaseEngine":
            return value
    return None


def get_configured_cache_instance_types(parsed_yamls):
    """
    Look up all the elasticache instance types referenced in the vars.
    """
    elasticache_instance_type_re = r"^cache.[a-zA-Z][\d]+[a-zA-Z]?\.\d*[a-zA-Z]+$"
    elasticache_instance_types = set()
    for data in utils.all_nested_values(parsed_yamls):
        if isinstance(data, six.text_type) and re.match(elasticache_instance_type_re, data):
            elasticache_instance_types.add(data)
    return set(elasticache_instance_types)


def get_configured_cache_engine(parsed_yamls):
    """
    Look up cache engine referenced in the vars.
    """
    for key, value in utils.all_nested_items(parsed_yamls):
        if key == "IncludeRedis" and value:
            return "Redis"
        elif key == "IncludeMemcached" and value:
            return "Memcached"
    return None


def get_configured_app_names(parsed_yamls):
    """
    Look up the app names configured in the vars.

    Returns:
        A pair where the first value is the frontend app name and the second value is the backend app name.
    """
    frontend_app_name = None
    backend_app_name = None
    for key, value in utils.all_nested_items(parsed_yamls):
        if key == "FrontendAppName":
            frontend_app_name = value
        if key == "BackendAppName":
            backend_app_name = value
    return frontend_app_name, backend_app_name
