# AWS GC (Garbage Collect)

This script can be used to garbage collect various AWS resources in the phxdevops test account that are not supported by
`cloud-nuke`.

Many of the resources in this script are not included for cleanup because it is hard to differentiate real resources
from test resources. This is primarily due to the nature of the resource. For example, in a sandbox AWS account, you
will still have real IAM roles and users because you need some credentials to be able to deploy and destroy resources.
`cloud-nuke` currently does not support any form of configuration involving regex or name globs for selecting resources,
which would be a prerequisite for adding these resources there. In the mean time, we use the scripts in this repo to
clean up resources that aren't supported by `cloud-nuke`.

Scripts:

- `aws-gc.py`: Clean up S3 buckets, IAM users, IAM groups, IAM instance profiles, IAM roles, AWS Config, and AWS
  Guardduty.
- `gc-ecs-cluster.py`: Clean up ECS clusters.
- `gc-ec2-instances.py`: Take termination protected EC2 instances and disable termination protection.


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
