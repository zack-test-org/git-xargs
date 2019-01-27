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
- `terragrunt update`: update the version number in a `.tfvars` file or even in a `module` in a `.tf` file. Perhaps you
  can use this to update multiple modules at once to some new version.
- `terragrunt price-estimate`: estimate how much a module will cost you to deploy.
