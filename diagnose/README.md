# Diagnose

`diagnose` is a prototype of a CLI tool that can be used to automatically diagnose connectivity issues. For example, 
let's say you deployed a web service at url `https://foo.example.com`, but it doesn't seem to be responding. What's the 
cause?

To answer the question, all you need to do is authenticate to AWS:

```bash
export AWS_ACCESS_KEY_ID=xxx
export AWS_SECRET_ACCESS_KEY=yyy
```

And run the `diagnose` utility:

```bash
diagnose https://foo.example.com
```

After a few seconds, it'll automatically figure out the problem:

```
======= DIAGNOSIS =======

None of the Security Groups attached to the ELB seem to allow outbound requests to the Instances. You should update 
those Security Groups on the ELB so it can send requests (including health checks) to your Instances! 
```

Like magic, this utility can automatically figure out if the issue is:

- The DNS entry isn't pointing at your ELB
- Your ELB health checks are failing
- Your ELB security group is misconfigured
- The security group on your EC2 Instances is misconfigured
- There's no process running on the instances and listening on your specified port
- (Many other types of checks can also be added in the future, such as NACLs, peering connections, etc)




## Building 

To build diagnose, install Go (at least `1.13`) and run:

```bash
go build -o /usr/local/bin/diagnose
```

This will create a `diagnose` binary in the same folder.




## Example

You can try this utility out with the code under `test/fixtures/asg-alb`. This Terraform module deploys an Auto Scaling
Group with a simple "Hello, World" web server, an ALB to route traffic across the ASG, and a Route 53 DNS entry 
pointing at the ALB:

```bash
cd test/fixtures/asg-alb
terraform init
terraform apply

...

Outputs:

url = http://jimtest.gruntwork.in
``` 

At the end of the `apply`, the example will output a URL you can use for testing:

```bash
$ curl http://jimtest.gruntwork.in
Hello, World!
```

So far, everything is working just fine... But now, let's try to break something! For example, a common error is to not 
add an ingress rule that allows the ELB to talk to the EC2 Instances. The `test/fixtures/asg-alb` module has a handy
`enable_broken_instance_security_group_settings` flag you can set to `true` try this very thing out!

```bash
terraform apply -var enable_broken_instance_security_group_settings=true
```

Now if you try to test the URL, you'll get no response for a while and then a 503 or 504 error from the ELB: 

```bash
$ curl http://jimtest.gruntwork.in
<html>
<head><title>504 Gateway Time-out</title></head>
<body bgcolor="white">
<center><h1>504 Gateway Time-out</h1></center>
</body>
</html>
```

The `diagnose` utility can figure out the problem automatically! First, authenticate to your AWS account:

```bash
export AWS_ACCESS_KEY_ID=xxx
export AWS_SECRET_ACCESS_KEY=yyy
```

Then run `diagnose`:

```bash
diagnose http://jimtest.gruntwork.in
```

After a few seconds, it should correctly diagnose the problem:

```
======= DIAGNOSIS =======

None of the Security Groups attached to the EC2 Instances seem to allow inbound requests from the ELB. You should 
update those Security Groups so the ELB can send requests (including health checks) to the apps running on those 
Instances!
```

You can experiment with other common errors too, such as omitting the egress rules on the ALB security group (which 
will prevent health checks from working):
 
```bash
terraform apply -var enable_broken_elb_security_group_settings=true
``` 
 
Re-run `diagnose`:

```bash
diagnose http://jimtest.gruntwork.in
```

And you'll see something like:

```
======= DIAGNOSIS =======

None of the Security Groups attached to the ELB seem to allow outbound requests to the Instances. You should update 
those Security Groups on the ELB so it can send requests (including health checks) to your Instances!
``` 
 
Or try updating the user data script so no web server actually runs on the EC2 instances:

```bash
terraform apply -var enable_broken_user_data=true
``` 

*(Note: for this change to take effect, you'll have to manually terminate the two EC2 Instances to force the ASG to 
deploy new ones).*

After running `diagnose`, you'll see:

```
======= DIAGNOSIS =======

Testing the instances via localhost failed. This most likely means your web service is not running or not listening on 
the port (8080) you expect.
```

Each time, the `diagnose` utility should tell you exactly what the issue is!




## How it works

The `diagnose` utility tries to automatically figure out what type of service you're running by using the AWS APIs. For
example, given a URL `foo.example.com`, the `diagnose` utility does the following:

1. Find the Route 53 Hosted Zone for `example.com`.
1. Find the Alias entry for `foo.example.com`.
1. Find the ELB that the Alias entry is pointing to.
1. Find all the Target Groups for that ELB.
1. Find all EC2 Instances in those Target Groups.
1. Check if the ELB Health Checks are passing for those Instances.
1. Check if the ELB Security Group allows outbound requests to those Instances.
1. Check if the Security Group for the Instances allows inbound requests from the ELB.
1. SSH to those Instances (via SSM) and see if there is a process locally listening on the right port.

If any of these checks fails, `diagnose` lets you know right away, as that's the best place to start your debugging!




## What's currently supported

`diagnose` was built as during the Gruntwork Hack Day of November, 2019, so the initially supported features are very,
very minimal. Currently, `diagnose` can only handle:

1. A Route 53 alias record...
1. Pointing to an ALB...
1. That is routing traffic to EC2 Instances...
1. That have an IAM role with SSM permissions. 

The code is a bit messy, and the `ShowDiagnosis` method, which shows a nice summary at the end, was only added at the
very last minute, and is not used everywhere it should be (e.g., Route 53 errors won't show up as nicely). PRs are 
very welcome :)




## What we could add in the future

The basic approach used here can be extended to support many, many other use cases:

1. ECS Services running in an ECS Cluster.
1. ECS Services running in Fargate.
1. K8S services running in EKS.
1. Lambda functions.
1. RDS connectivity.
1. ElastiCache connectivity.
1. VPC checks: e.g., validate NACLs, peering connections, etc.
1. Visualizations: show the connectivity graph (e.g., Route 53 -> ELB -> Instances) and which checks are passing and 
   failing visually.
1. And so on.



 