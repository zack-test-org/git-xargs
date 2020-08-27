import urllib.request
import boto3
from urllib.error import HTTPError
import json
import os


def make_github_request(url, github_token):
    headers = {
        'Authorization': f'token {github_token}'
    }

    req = urllib.request.Request(url, headers=headers)

    try:
        response = urllib.request.urlopen(req)
    except HTTPError as e:
        return e.code, None

    status_code = response.getcode()
    body = response.read()

    return status_code, body


def format_response(body, status_code=200):
    body_as_json = json.dumps(body) if not isinstance(body, str) else body

    return {
        'statusCode': status_code,
        'body': body_as_json
    }


def get_github_oauth_token(aws_region, secret_name):
    token = os.environ.get('GITHUB_OAUTH_TOKEN')
    if token:
        print(f'Using GitHub token from environment variable GITHUB_OAUTH_TOKEN')
        return token

    print(f'Did not find GitHub token in environment variable GITHUB_OAUTH_TOKEN, so looking it up in AWS Secrets Manager in region ${aws_region} with ID ${secret_name}')

    # Create a Secrets Manager client
    session = boto3.session.Session()
    client = session.client(
        service_name='secretsmanager',
        region_name=aws_region
    )

    response = client.get_secret_value(
        SecretId=secret_name
    )

    return response['SecretString']
