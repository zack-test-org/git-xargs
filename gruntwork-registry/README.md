# Gruntwork Registry Prototype

This is a quick and hacky prototype of the Gruntwork Registry. The only thing it can do right now is serve an endpoint
for [RenovateBot](https://renovate.whitesourcesoftware.com/) that returns the available versions for our modules. This
gives us auto-update functionality in the [aws-service-catalog 
repo](https://github.com/gruntwork-io/aws-service-catalog/).




## Quick start

This app is built using the [Serverless Framework](https://www.serverless.com/). To run it locally:

1. [Install Node.js](https://nodejs.org/en/download/).

1. [Install the Serverless Framework](https://www.serverless.com/framework/docs/getting-started/).

1. Run `npm install`.

1. Export a [GitHub personal access 
   token](https://docs.github.com/en/github/authenticating-to-github/creating-a-personal-access-token):
   
    ```bash
    export GITHUB_OAUTH_TOKEN=xxx
    ```
    
1. Run the app in offline mode:

    ```bash
    serverless offline
    ```
    
1. Try out various endpoints:

    1. Test the service discovery endpoint: http://localhost:3000/default/.well-known/terraform.json        
    1. Get all the available versions for `module-security`: http://localhost:3000/default/v1/modules/gruntwork-io/module-security/aws
    



## Deploy

We deploy this app officially to the Gruntwork sandbox account (see [Gruntwork AWS 
accounts](https://www.notion.so/AWS-Accounts-d936fc8f10674c9aafef34c4de87f2f2)). 

1. Authenticate to the sandbox account.

1. Deploy:

    ```bash
    serverless deploy
    ```           

1. This should deploy the app to AWS and configure it with the domain name `registry.gruntwork-sandbox.com`. You can 
   try out various endpoints:     

    1. Test the service discovery endpoint: https://registry.gruntwork-sandbox.com/.well-known/terraform.json        
    1. Get all the available versions for `module-security`: https://registry.gruntwork-sandbox.com/v1/modules/gruntwork-io/module-security/aws




## Secrets management

This app requires a GitHub personal access token to be able to call the GitHub API and fetch version information for our
private repos. 

- In dev, you export your own personal access token using the `GITHUB_OAUTH_TOKEN` environment variable.

- In AWS, someone manually needs to put a machine user's personal access token into AWS Secrets Manager into the same
  region as the app (this has already been done in the sandbox account). See [`serverless.yml`](serverless.yml) for 
  the AWS region and [`handler.py`](./src/handler.py) for the Secrets Manager ID. Note that 
  [`serverless.yml`](serverless.yml) also attaches permissions to the IAM role to read that secret from Secrets Manager.
    