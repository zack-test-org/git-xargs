# Recreate Default VPCs

This recreates the default VPC in all enabled regions of the authenticated account. This is useful when testing CIS
compliance features, where we have to delete the default VPC in all enabled regions.

## Usage

```
pip install -r requirements.txt
# Authenticate to the target account
python recreate_default_vpcs.py
```
