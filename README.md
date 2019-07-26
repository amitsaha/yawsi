## Yet Another AWS Command Line Interface

This is `yawsi` - a hobby project to implement a very minimal
subset of functionalities usually offered by AWS command line clients. It also has a set
of commands directly exposing certain "workflow" scenarios that you would need to
look up the AWS CLI manual for. 

### Top reasons why you may be interested

I think the commands I am most excited about are and perhaps worth your time are:

- [yawsi ec2 inspect](./docs/yawsi_ec2_inspect.md)
- [yawsi ec2 launch-more-like](./docs/yawsi_ec2_launch-more-like.md)

Some of the sub-commands provide an fuzzy interactive interface making use of 
[go-fuzzyfinder](https://github.com/ktr0731/go-fuzzyfinder) which is another feature I really like.

The command [yawsi vpc list-nacl-entries](./docs/yawsi_vpc_list-nacl-entries.md) also has the capability
to generate [Terraform](ttps://www.terraform.io/) code for AWS network ACL entries given an AWS network acl ID.
I found this to be really useful when importing existing AWS NACL resources into Terraform.

For a list of all the commands/sub-commands, please see [docs](./docs/yawsi.md).


## Install

Binary releases are available from the [releases](https://github.com/amitsaha/yawsi/releases) page. 
Please download the ZIP corresponding to your OS/architecture, unzip the file and place the binary somewhere
on your system which is added to the system path. 

### Bash completion

To get automatic Tab completion of the commands, options and flags, put `complete -C yawsi yawsi` somewhere
in your `~/.bashrc`.


## Specifying AWS profile

Specify the AWS profile via the `AWS_PROFILE` environment variable. You may also need to specify
the `AWS_REGION` environment variable explicitly. See [here](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html) to learn more and other
options.

Example setup:

```

1. Create ~/.aws/credentials of the form:

 [profile_name]
 aws_access_key_id=
 aws_secret_access_key=
 ..

 2. Create ~/.aws/config of the form:
 [profile profile_name]
 region=ap-southeast-2/us-east-1

 ```


## Building the binary

You will need `go 1.12+` installed:

```
$ make build BINARY_NAME=yawsi
```

## Building DEB packages

Building on your host system is supported, but you will need
[fpm](https://github.com/jordansissel/fpm) installed:

```
$ make build-deb DEB_PACKAGE_DESCRIPTION="Yet another AWS CLI" DEB_PACKAGE_NAME=yawsi BINARY_NAME=yawsi HOST_BUILD=yes
```

Or if you have `docker` installed and configured to be usable as
normal user:

```
$ make build-deb DEB_PACKAGE_DESCRIPTION="Yet another AWS CLI" DEB_PACKAGE_NAME=yawsi BINARY_NAME=yawsi
```

In both cases, the resulting DEB package is in `artifacts` directory.

## License

See `LICENSE`.
