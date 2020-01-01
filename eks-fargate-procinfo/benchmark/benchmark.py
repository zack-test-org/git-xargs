import os
import subprocess
import re
import boto3
import uuid
import timeit
import gzip

GZJSON_FILE = '/tmp/scripts/data.json.gz'
with gzip.open(GZJSON_FILE) as f:
    JSON_STR = f.read()
REGION = os.environ['REGION']
BENCHMARK_TABLE_NAME = os.environ['BENCHMARK_TABLE_NAME']

dynamodb = boto3.client('dynamodb', region_name=REGION)


def get_processor_name():
    command = 'cat /proc/cpuinfo'
    all_info = subprocess.check_output(command, shell=True).decode().strip()
    for line in all_info.split('\n'):
        if 'model name' in line:
            return re.sub('.*model name.*:', '', line, 1)
    raise Exception('Could not determine processor type')


def load_json():
    import json
    json.loads(JSON_STR)


def benchmark():
    timer = timeit.Timer('load_json()', globals=globals())
    five_runs = timer.repeat(repeat=5, number=1000)
    best_of_five = min(five_runs)
    return best_of_five


def run():
    id_ = str(uuid.uuid4())
    procinfo = get_processor_name()
    runtime = benchmark()
    dynamodb.put_item(
        TableName=BENCHMARK_TABLE_NAME,
        Item={
            'UUID': {
                'S': id_
            },
            'ProcInfo': {
                'S': procinfo
            },
            'RunTime': {
                'N': str(runtime)
            },
        }
    )


if __name__ == '__main__':
    run()
