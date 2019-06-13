# Terraform On Rails

**TODO: Turn this into a proper RDD README**

From: https://gruntwork-io.slack.com/archives/C0PJF332B/p1547645541002900

A while back, I had mentioned the idea that we could create “Terraform on Rails.” That is, in the same way that Ruby on
Rails gives you a streamlined, opinionated, batteries-included way to build web apps with Ruby, we could do the same for
developing infrastructure with Terraform. I wanted to write down some thoughts I had about that, as well as a new way to
seriously simplify testing with Terratest.

Let’s imagine that Terragrunt 1.0.0 (or 2.0.0?) is the tool we use to power Terraform on Rails. It doesn’t have to be
Terragrunt, but we’ll need some CLI tool similar to the `rails` CLI, so let’s assume that’s Terragrunt for now. Here’s
the new way you’d use Terraform/Terragrunt:

- To create a new module from scratch, you’d run `terragrunt new module`.
	- This would prompt you for some info about the module—e.g., its name, what
      to use as a remote state backend, what env to deploy it to, etc—and create our
      opinionated file/folder layout.
    - The file/folder layout may change in the future, but let’s assume for now that we’ll stick with the
      `infrastructure-live` / `infrastructure-modules` split.
    - `terragrunt new` would generate:
        - `infrastructure-modules/<your-module>` with contents `main.tf` (including the `terraform { }` block with
          version pinning and `backend "<...>" { }` config and some dummy resource), `variables.tf` (with a dummy
          variable called, say, `example_var`), `outputs.tf` (with some dummy output), and `dependencies.tf`.
        - `infrastructure-modules/test/<your-module>.go` with a skeleton of an automated test for the module (which
          actually deploys that module, tests its outputs, and then destroys it!).
        - `infrastructure-live/<some-env>/<your-module>/terraform.tfvars` with the proper remote state config, `include`
          settings, a dummy value for `example_var`, etc.

    - The goal is to immediately be able to run `terragrunt apply` to deploy that module and to run `terragrunt test` to
      run the tests for that module.

- More generally, we could have a `terragrunt new <URL>` or similar command:
    - `URL` is a URL to a Git repo that contains Terraform code.
    - More specifically, that repo could contain _templated_ Terraform code. We probably wouldn't use `boilerplate` for this, but it gets the idea across well, so let's assume `boilerplate` for the purposes of this doc.
    - We would create our module repos (e.g., `module-vpc`, `terraform-aws-eks`, etc) with our usual folder structure:`modules`, `examples`, and `test`.
    - We could add a new folder to this structure called something like `templates`. This would contain subfolders with templated Terraform code.
    - E.g., `module-vpc/templates/prod-app-vpc` could contain:
        - `boilerplate.yml`: defines input variables to ask the user for, such as the name of the VPC and CIDR block to use.
        - `infrastructure-modules`: would contain `main.tf`, `variables.tf`, etc for a VPC module. These would use Go templating syntax to fill in the boilerplate variables the user asked for.
        - `infrastructure-live`: would contain `terragrunt.hcl`, also with Go templating syntax.
    - The `terragrunt new <URL>` command would check out the code at `URL` and run `boilerplate` on it. The user would be prompted to enter input variables and based on those, we'd generate a new module for them.
    - This would provide a standardized way to create "scaffolding" for Terraform projects. Using this, it would be 10x easier for customers to try out parts of the IaC Library: that is, we could turn all the contents of `usage-patterns` into a service catalog of reusable scaffolds so that customers can create a production-grade ref arch for themselves, piece by piece. We could have various flavors of scaffolding, such as an app VPC, PCI-compliant EKS deployment, frontend service for deployment into EKS, etc. 
    - These same scaffolds could be reused by Houston self service to offer a UI-driven service catalog. 
  
- We could add new `terragrunt release` and `terragrunt promote` commands to help with day to day operations.
    - `terragrunt release` would create a new Git tag in your `infrastructure-modules` repo.
    - `terragrunt promote` would know how to update `terragrunt.hcl` files  to promote a new version of some module from dev to stage to prod:
        - Update the `ref=` parameter in `terragrunt.hcl`
        - Commit the changes to Git
        - Run `terragrunt apply`

- I think we could have a more streamlined test structure for Terratest tests too. We could add a wrapper to Terratest
  to significantly simplify testing Terraform modules. Something along the lines of:
    ```
    func TestModule(t *testing.T, options *terraform.Options, validate func(outputs *terraform.Outputs))
    ```
    - When you call this method, you pass it the path to your module and the variables to set for it, and the method, on
      your behalf:
        - Runs `terraform init`
        - Runs `terraform apply`, forwarding your vars along
        - Calls your `validate` function, passing it all the outputs from your module
        - Calls `terraform destroy` at the end of the test

    - So then your test code reduces to something like:
        ```
        func MyTest(t *testing.T) {
          options := &terraform.Options {
            Path: "../my-module",
            Vars: map[string]interface{
              "foo": "bar"
            },
          }
        
          terraform.TestModule(t, options, func(outputs *terraform.Outputs) {
            publicIp := outputs.get("public_ip") as string
            http_helper.HttpGetWithRetry(t, publicIp)
          })
        }
        ```

- `terragrunt new project`: create a totally new project, including `infra-live` and `infra-modules` repos and root
  `terraform.tfvars` files.
- `terragrunt new module`: create a new module within an existing project.
- `terragrunt price-estimate`: estimate how much a module will cost you to deploy.
