#!/bin/sh
#
# Entrypoint script to run the benchmark code from a bare python docker container. This involves installing the required
# python requirements before calling the script.
#

pip install boto3
python /tmp/scripts/benchmark.py
