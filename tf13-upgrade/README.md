# Terraform 0.13 upgrade script

This is a script that automates some of the steps for the [Terraform 0.13 
upgrade](https://www.notion.so/gruntwork/Terraform-0-13-Upgrade-0c88a38ab19e4f588d253c1733259bcd):

1. Automatically find all folders with Terraform code (`*.tf`) in them
1. In each folder:
    1. Run `terraform 0.13upgrade`.
    1. Relax the `required_version` constraint in `versions.tf` to support Terraform 0.12.
    1. Look for duplicate `required_version` usage.
    1. Look for destroy-time `provisioner` usage.
1. Print instructions on next steps.



## Usage

See [Terraform 0.13 Upgrade](https://www.notion.so/gruntwork/Terraform-0-13-Upgrade-0c88a38ab19e4f588d253c1733259bcd)
for full instructions. 

This script takes only a single, optional argument:

```bash
./tf13-upgrade.sh [PATH]
```

`PATH` is the folder where to run this script. It defaults to the current working directory. 
 