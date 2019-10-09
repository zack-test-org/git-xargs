# Fork repos

This folder contains two Bash scripts that can be used to fork (i.e., copy) the Git repos in the [Gruntwork 
Infrastructure as Code Library (IaC Library)](https://gruntwork.io/infrastructure-as-code-library/) into your company's 
private Git repos, including updating all internal and cross-references in the code to also point to your company's Git 
repos, so you can use the IaC Library completely from within your company's network and repositories, with no external
links to Gruntwork's GitHub accounts. These scripts are designed to be idempotent, so you can run them on a scheduled 
basis (e.g., as part of a cron job) to periodically pull in the latest code from the Gruntwork repos.

The two scripts are:

1. `fork-repo.sh`: Fork a single repo and update all cross-references. Use this script if you need to fork just one 
   repo or you are just experimenting with the scripts.
1. `fork-all-repos.sh`: Loop over all of Gruntwork's repos and use `fork-repo.sh` to fork each one. Use this script if
   you want to fork the entire IaC Library. 

In the next two sections, we'll describe why we have special scripts for forking the code and what the scripts do and
don't do, and in the two sections after that, we'll describe how to use the scripts.




## Why we need these scripts and what they do

### The problem

Normally, copying code from one Git repo to another is a simple process: you `git clone` one repo, add a new 
destination using `git add remote`, and then `git push` to that remote. Unfortunately, copying the IaC Library is more
complicated because:

1. Some of the repos have _cross-references_ to other repos. For example, one of the Terraform modules in the 
   `package-kafka` repo uses a Terraform module from the `module-asg` repo to spin up an Auto Scaling Group that can
   support persistent EBS Volumes and ENIs:
    ```hcl
    module "kafka_brokers" {
      source = "git::git@github.com:gruntwork-io/module-asg.git//modules/server-group?ref=v0.8.1"
      # ...
    }
    ```
    Terraform does not allow the use of variables in `source` URLs, so there is no way to make those URLs parameterized.    
    Therefore, to be able to use all the code from the Gruntwork IaC Library solely from your own repos, all of these 
    cross-references need to be changed, directly in the code, to point to your own repos.
    
1. The cross-references are _versioned_. For example, the`package-kafka` code above references `v0.8.1` of `module-asg`,
   which is a Git tag in the `module-asg` repo. But `module-asg` might have its own cross-references to still other
   Gruntwork repos, and after we've updated those cross-references, we need to publish a new tag with those changes.
   So you not only need to update the cross-reference URLs, but also the version numbers they use.     

1. Not every cross-reference is pointing to the latest version of a repo. For example, the `package-kafka` code above
   references `v0.8.1` of `module-asg`, but it's possible that there is already a `v0.9.0` of `module-asg` available.
   Since the newer versions could have backwards incompatible changes, you can't just update cross-references in the 
   latest version of `master`, release a new tag, and update all other modules to use that new tag—you actually have to
   update all old tags too!
   
### The solution
   
Here's how the two scripts in this folder solve this messy problem:  

1. The `fork-all-repos.sh` script loops over each Gruntwork repo `gruntwork-io/foo` and calls the `fork-repo.sh` 
   script, telling it to make a copy of `gruntwork-io/foo` in some repo `<your-company>/foo`.
1. The `fork-repo.sh` script does the following:
    1. `git clone` the `gruntwork-io/foo` repo into a temp folder.
    1. Loop over each tag `vX.Y.Z` in `gruntwork-io/foo`, one at a time.
    1. Create a new branch called `vX.Y.Z-internal` from tag `vX.Y.Z`. 
    1. Update all cross-references as follows:
        1. Update all `gruntwork-io/xxx` URLs to point to `<your-company>/xxx`.
        1. Update all references/tags of the form `vA.B.C` to point to `vA.B.C-internal`. 
    1. Commit the changes to branch `vX.Y.Z-internal`.
    1. `git push` the branch `vX.Y.Z-internal` to `<your-company>/foo`.
    
The result of running these scripts is that every cross-reference to version `vX.Y.Z` in the Gruntwork repos will be
updated to a cross-reference to version `vX.Y.Z-internal` in your company's repos. Since the script creates a branch of
exactly this `vX.Y.Z-internal` name in each repo, all the cross-references should work exactly as expected!
     
Note that (a) the script performs the same process on the `master` branch, so that you have a reasonable deafult branch 
in `<your-company>/foo` and (b) the script is idempotent, so if branch `vX.Y.Z-internal` or `master` already exists in 
`<your-company>/foo`, the script will not update that branch or push new code to it again.


  

## What the scripts don't do

These scripts are not perfect. They are a best effort to solve a tricky problem. Here are some known limitations:

1. **Some cross-references may not be updated perfectly**. These scripts use `grep` and `sed` to make a best effort at 
   updating all cross-references, but updating code with search and replace is an imperfect process, and it's possible
   some cross-references will be missed, or something will accidentally be updated that shouldn't have been. If you 
   find such an issue, please let us know!

1. **Assets**: These scripts do NOT copy published assets from the Releases page of each repo. For example, some of or
   modules are written in Go, and as part of the release process, we publish pre-compiled, standalone binaries for each
   OS (e.g., for the `gruntkms` repo, we publish `gruntkms_linux_amd64`, `gruntkms_darwin_amd64`, 
   `gruntkms_windows_amd64.exe`, etc.). These binaries are published as GitHub release assets, so they are not in the
   Git repo itself, and will NOT be copied by these scripts. If you need these binaries, you will need to either copy
   them manually and/or build them directly from source. 




## How to use `fork-repo.sh`

If you want to fork just a single repo, you can run `fork-repo.sh`. To fork a Gruntwork repo `gruntwork-io/foo`:

1. Make sure you have `git clone` access to `gruntwork-io/foo` on whatever computer you'll be using to run 
   `fork-repo.sh`. See [Get access to the Gruntwork Infrastructure as Code 
   Library](https://gruntwork.io/guides/foundations/how-to-use-gruntwork-infrastructure-as-code-library/#get_access)
   for instructions. 
1. Create a Git repo `<your-company>/foo` in your company's Git repos where you can push the forked code.
1. Run `fork-repo.sh [ARGUMENTS]`, where the supported arguments are:

| Argument          | Required | Description                                                                                                                                                                             |
|-------------------|----------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `--src`           | Yes      | The URL of the Gruntwork repo to fork. This script will git clone this repo.                                                                                                            |
| `--dst`           | Yes      | The URL of your repo. This script will push the forked code here.                                                                                                                       |
| `--base-https`    | Yes      | The base HTTPS URL for your organization. E.g., https://github.com/your-company. This is used to replace https://github.com/gruntwork-io URLs in all cross-references.                  |
| `--base-git`      | Yes      | The base Git URL for your organization. E.g., git@github.com:your-company or github.com/your-company. This is used to replace git@github.com:gruntwork-io URLs in all cross-references. |
| `--dry-run`       | No       | If this flag is set, perform all the changes locally, but don't push them to the `--dst` repo. This will leave the temp folder on disk so you can inspect what would've been pushed.    |
| `--dry-run-local` | No       | Same as `--dry-run`, but also skip fetching data from the destination repo or checking if branches already exist. This lets you test locally without creating a destination repo.       |
| `--help`          | No       | Show help text and exit.                                                                                                                                                                |

For example, to fork the `gruntwork-io/module-vpc` repo into a GitLab repo called `acme-corp/module-vpc`, you would
run: 

```bash
./fork-repo.sh \
  --src git@github.com:gruntwork-io/module-vpc
  --dst git@gitlab.com:acme-corp/module-vpc.git \
  --base-https https://gitlab.com/acme-corp/module-vpc \
  --base-git git@gitlab.com:acme-corp/module-vpc.git
```

If `acme-corp/module-vpc` doesn't exist yet, and you just want to experiment and see the updates this script would make, 
you can add the `--dry-run-local` flag:

```bash
./fork-repo.sh \
  --src git@github.com:gruntwork-io/module-vpc
  --dst git@gitlab.com:acme-corp/module-vpc.git \
  --base-https https://gitlab.com/acme-corp/module-vpc \
  --base-git git@gitlab.com:acme-corp/module-vpc.git \
  --dry-run-local
```

The script will `git clone` the `module-vpc` repo into a temp folder, update all the cross references, create new 
branches for all the release tags, and commit the changes locally into the temp folder. At the end of the script, it 
will print out the path of the temp folder and all the branches that were created so you can inspect the resulting code.  




## How to use `fork-all-repos.sh`

If you want to fork the entire Gruntwork IaC Library, you can run `fork-all-repos.sh`:

1. Open `fork-all-repos.sh` and check out the `GRUNTWORK_REPOS` and `GRUNTWORK_HASHICORP_REPOS` variables to see what
   repos will be forked. Feel free to comment repos out in these two variables if you only need a subset of the repos.
1. Make sure you have `git clone` access to all the repos in `GRUNTWORK_REPOS` and `GRUNTWORK_HASHICORP_REPOS` on 
   whatever computer you'll be using to run `fork-all-repos.sh`. See [Get access to the Gruntwork Infrastructure as Code 
   Library](https://gruntwork.io/guides/foundations/how-to-use-gruntwork-infrastructure-as-code-library/#get_access)
   for instructions. 
1. For each repo `gruntwork-io/foo` in `GRUNTWORK_REPOS` and `GRUNTWORK_HASHICORP_REPOS`, create a Git repo 
   `<your-company>/foo` in your company's Git repos where you can push the forked code.
1. Run `fork-all-repos.sh [ARGUMENTS]`, where the supported arguments are:
 
| Argument          | Required | Description                                                                                                                                                                                 |
|-------------------|----------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `--base-https`    | Yes      | The base HTTPS URL for your organization. E.g., https://github.com/your-company. This is used to replace https://github.com/gruntwork-io URLs in all cross-references.                      |
| `--base-git`      | Yes      | The base Git URL for your organization. E.g., git@github.com:your-company or github.com/your-company. This is used to replace git@github.com:gruntwork-io URLs in all cross-references.     |
| `--suffix`        | No       | If specified, this suffix will be appended to every repo name. That is, each Grunwork repo `foo` will be pushed to an internal repo of yours called `foo<suffix>`. Default: (empty string). |
| `--dry-run`       | No       | If this flag is set, perform all the changes locally, but don't git push them. This will leave the temp folders on disk so you can inspect what would've been pushed.                       |
| `--dry-run-local` | No       | Same as `--dry-run`, but also skip fetching data from the destination repo or checking if branches already exist. This lets you test locally without creating a destination repo.           |
| `--help`          | No       | Show help text and exit.                                                                                                                                                                    |

Example usage: 

```bash
./fork-all-repos.sh \
  --base-https https://gitlab.com/acme-corp/module-vpc \
  --base-git git@gitlab.com:acme-corp/module-vpc.git
```

If the destination repos doesn't exist yet, and you just want to experiment and see the updates this script would make, 
you can add the `--dry-run-local` flag:

```bash
./fork-all-repos.sh \
  --base-https https://gitlab.com/acme-corp/module-vpc \
  --base-git git@gitlab.com:acme-corp/module-vpc.git \
  --dry-run-local
```

The script will `git clone` each repo into a temp folder, update all the cross references, create new branches for all 
the release tags, and commit the changes locally into the temp folder. After processing each repo, it will print out the 
path of the temp folder and all the branches that were created so you can inspect the resulting code.  
