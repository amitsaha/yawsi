## Yet Another AWS Command Line Interface

This is `yawsi` - a hobby project to implement a very minimal
subset of functionalities usually offered by AWS command line clients.

## Install

Binary releases are available from the Releases page. Please download the ZIP corresponding to your OS and architecture and unzip the binary somewhere in your $PATH.


## Specifying AWS profile

Specify the AWS profile via the `AWS_PROFILE` environment variable.
Example setup:

```

1. Create ~/.aws/credentials of the form:

 [profile_name]
 aws_access_key_id=
 aws_secret_access_key=
 ..

 2. Create ~/.aws/config of the form:
 [profile_name]
 region=ap-southeast-2/us-east-1

 ```

## Sub-commands

All current functionalities currently are available via the `ec2` sub-command:

```
$ AWS_PROFILE=dev yawsi
Yet Another AWS Command Line Interface

Usage:
  yawsi [command]

Available Commands:
  ec2         Commands for working with AWS EC2
  help        Help about any command

Flags:
  -h, --help   help for yawsi

Use "yawsi [command] --help" for more information about a command.
```

EC2 sub-commands:

```

$ AWS_PROFILE=dev yawsi ec2 help
Commands for working with AWS EC2

Usage:
  yawsi ec2 [command]

Available Commands:
  launch-more-like Launch AWS EC2 classic instance like another instance
  list-asgs        List Autoscaling Groups
  list-instances   List EC2 instances

```

## Examples

**List all EC2 instances**

```
$ yawsi ec2 list-instances
i-031a7bbcfb163de12 : running : 127h8m23.809358629s : ec2-54-206-131-205.ap-southeast-2.compute.amazonaws.com : ip-10-219-32-208.ap-southeast-2.compute.internal
...

```

**List all EC2 instances having certain tags**

```
$ yawsi ec2 list-instances --tags "key1:value1,key2:value2"
...
```

**List all EC2 instances attached to an autoscaling group**

```
$ yawsi ec2 list-instances --asg myasgname
...
```

**Launch an EC2 instance copying the configuration from another EC2 instance**

```
$ yawsi ec2 launch-more-like <instance-id>
```

**Launch an EC2 instance copying the configuration from another EC2 instance, but modifying the user data**

```
 $ AWS_PROFILE=dev go run main.go ec2 launch-more-like <instance-id> --edit-user-data
```

It looks up the `EDITOR` environment variable to find out the editor that you will be using
to edit the user data. If one is not found, it defaults to `vim`. You can override it as:

```
$ EDITOR=nano ..
```

**List all auto scaling groups**

```
$ yawsi ec2 list-ags
```
## Building the binary

You will need `go 1.8+` installed:

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
