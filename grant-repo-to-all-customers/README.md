# Grant repo to all customers

This script can be used to grant all our customers access to the specified repo. The main use case is if there is a new
repo that was added to the library and you want to retroactively grant access to the repo to all the customers.

*Note*: This script requires python 3.6+

## Usage

The script works by using a seed repo as a reference for knowing which customers to grant access to. For example, if you
want to only grant access to customers with CIS subscription, you would set the seed repo to a repo that is only
available in the CIS subscription (e.g `cis-compliance-aws`).

```bash
export GITHUB_OAUTH_TOKEN=xxx
pip install -r ./requirements.txt
python ./grant_to_all_customers.py --repo new-repo --seed cis-compliance-aws
```

## Which seed repo should I use for each subscription?

**aws**: `module-vpc`
**gcp**: `terraform-helm-gke-exts`
**aws-cis**: `cis-compliance-aws`
**enterprise**: `module-ci`
