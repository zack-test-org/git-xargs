import boto3
import json
import logging
import pkg_resources

from . import utils


# Global objects. Should only include objects that should only be defined once in the script

## logger for the script
logger = utils.get_configured_logger()
LOG_LEVEL_MAP = {
    "debug": logging.DEBUG,
    "info": logging.INFO,
    "warn": logging.WARNING,
    "error": logging.ERROR,
}

## boto3 client to access STS API
sts_client = boto3.client("sts")

## boto3 client to access pricing API
pricing_client = boto3.client("pricing", region_name="us-east-1")

## boto3 endpoint list. Used to map region codes to descriptions (pricing API uses descriptions)
aws_endpoint_file = pkg_resources.resource_filename("botocore", "data/endpoints.json")
with open(aws_endpoint_file) as f:
    aws_endpoints = json.load(f)
