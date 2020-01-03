"""
Benchmark script to collect information about the underlying processor and record it in a specified DynamoDB table. The
underlying CPU benchmark is one of the standard JSON parsing test used in the nativejson-benchmark
(https://github.com/miloyip/nativejson-benchmark).

Expected environment variables:
- REGION : The AWS Region where the script is run
- BENCHMARK_TABLE_NAME : The name of the DynamoDB table to store results in

Expected Files:
- /tmp/scripts/data.json.gz : A gunzip compressed json file used in the benchmark.
"""

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
    """
    Return the model name of the underlying CPU, collected from /proc/cpuinfo. Raises an exception if the model could
    not be determined from /proc/cpuinfo.
    """
    command = 'cat /proc/cpuinfo'
    all_info = subprocess.check_output(command, shell=True).decode().strip()
    for line in all_info.split('\n'):
        if 'model name' in line:
            return re.sub('.*model name.*:', '', line, 1)
    raise Exception('Could not determine processor type')


def load_json():
    """
    Benchmark function. This routine will be executed multiple times to collect CPU runtime information.
    """
    import json
    json.loads(JSON_STR)


def benchmark():
    """
    Using timeit, repeatedly call the load_json benchmark.

    Each run collects overall runtime of calling load_json 1000 times. This is repeated 5 times to collect 5 runtimes.
    The best run of the 5 is ultimately returned.

    See the timeit docs for more information: https://docs.python.org/3/library/timeit.html
    """
    timer = timeit.Timer('load_json()', globals=globals())
    five_runs = timer.repeat(repeat=5, number=1000)
    best_of_five = min(five_runs)
    return best_of_five


def run():
    """
    Main routine. Collect processor information, run the benchmark, and then record the results in the DynamoDB table.
    """
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
