# Secure Secrets Provider Prototype

This folder contains prototype code for a custom Terraform provider that:

1. Can store secrets in AWS Secrets Manager...
1. Without those secrets ending up in Terraform state.

I think the approach used in this provider could be used to also support creating other types of secrets that we want 
to keep out of Terraform state: e.g., self-signed TLS certs.

This code was originally created as a Hack Day project and is NOT production ready. However, with not too much more 
work, we could turn it into a real provider, publish it in the Terraform Registry, and start using it to manage secrets
more securely, while still getting all the benefits of integration with Terraform and its lifecycle.

For more context, see:

1. [Video walkthrough of using this provider](https://gruntwork-io.slack.com/archives/CM6RR6JE7/p1606062990018000).
1. [Video walkthrough of the code](https://gruntwork-io.slack.com/archives/CM6RR6JE7/p1606067481020600).



## Using the provider

1. Make sure you have Go >=1.13 installed.
1. Run `make install`.
1. `cd examples/simple`.
1. Authenticate to AWS on the CLI.    
1. `terraform init`.
1. `terraform apply`.
1. `terraform destroy`.

Every time you make a change to the provider code, re-run `make install` and `terraform init` to see your latest 
changes.




## Debugging the provider

The best way I've found to debug so far is to use logging. Update the code with some `log.Printf(...)` statements 
(using the `log` package built into Go), run `make install`, `terraform init`, and before running `apply`, enable
[Terraform logging](https://www.terraform.io/docs/internals/debugging.html):

```bash
export TF_LOG=debug
terraform apply
```

You'll get a lot of extra log output this way, but if you dig thorugh it, you'll find your `log.Printf(...)` statements
too.



## TODOs

1. Add `ForceDeleteWithoutRecovery` as a param.
1. Add tests. See [Terraform provider unit tests](https://www.terraform.io/docs/extend/testing/unit-testing.html) and
   [Terraform provider acceptance tests](https://www.terraform.io/docs/extend/testing/acceptance-tests/index.html).
1. Add more example usage.
1. Rename the provider and resources to use `gruntwork_<cloud>_<provider>_<resource_name>` as the naming convention. 
   E.g., The resource to store secrets should be called `gruntwork_aws_secret_value`.   
1. Look into the existing Terraform `aws_secretsmanager_xxx` resources and see if: 
    1. We should follow their naming conventions so our resources are an analogous but more secure way to do the same 
       thing, making it easier to switch.
    1. If we should support the other params those resources do, such as tagging, versioning, etc.
1. Add support for creating and storing self-signed TLS certs in AWS Secrets Manager, without  the private keys ending
   up in Terraform state. 
    1. NOTE: instead of writing all the TLS cert code from scratch, see the existing Terraform `tls`
       provider code, and see if we can re-use that! That code already generates certs just fine, so the only change is to
       keep them out of state.
    1. NOTE: the `tls` provider and most other HashiCorp code is typically MPL licensed. We are adding that as an 
       accepted license, but we need to make sure that process is done first before using the library.              


