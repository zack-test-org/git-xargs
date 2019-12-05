# AWS GC (Garbage Collect)

This script can be used to garbage collect IAM resources and S3 buckets in the phxdevops test account.

We often hit the IAM and S3 limits due to test failures. This is problematic because cloud-nuke does not clean up IAM
resources and S3 buckets due to the difficulty in differentiating real resources from test resources. This script
relies on the operator to scan the list it finds, making sure there aren't any resources that it shouldn't delete.

## Usage

This script depends on a few python libraries. Install them using pip:

```
pip install -r requirements.txt
```

Once the dependencies are installed, you can run the script using python:

```
python aws-gc.py
```

By default the script will run in dry mode, only reporting the resources that it finds that fit the deletion criteria.
To actually delete the resources, pass in the `-r` arg:

```
python aws-gc.py -r
```

*Note*: This script requires python 3.6+. The `python` binary on your local machine may be named `python3`.
