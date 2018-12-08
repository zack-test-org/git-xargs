# Reference Architecture Sanity Check

This is a python script that will parse the boilerplate input YAML file and sanity checks the inputs to help catch
potential issues before you proceed with the reference architecture deployment.

The following checks are implemented:

- Check that you can access each customer account from the Gruntwork Customer Access account.
- Check that all of the Acme specific vars are deleted.
- Check that the sample app names are different
- Check that the ECR repository name is configured properly.
- Check that the accounts have route 53 domains for the defined domains.
- Check that ACM certificates exist for all the domains.
- Check that all the instance types specified for EC2, RDS, and ElastiCache exist in the region.

The following checks are suggested for future improvement:

- Check that db instance supports end to end encryption.


## Usage

- Make sure python is available
- Install requirements (`pip install -r requirements.txt`)
- Go to your `usage-patterns` base directory
- Authenticate to the customers' security account. Make sure you are MFA authenticated in the CLI.
- Run the `refarch-sanity-check` script and pass in the customer name you want to test. For example:

```
$PATH_TO_PROTOTYPES_REPO/refarch_sanity_check/refarch-sanity-check --customer-name acme-multi-account
```

## Graduation thoughts

When this script has been tested across a few reference architecture deployments, we should consider graduating the
script to [`usage-patterns`](https://github.com/gruntwork-io/usage-patterns) directly. At that point, this script should
be called automatically by boilerplate as part of a `pre` hook when `IsLegacyCustomerOrTest` is not true.
