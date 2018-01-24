## Yet another AWS CLI

### Install

```
$ go get github.com/amitsaha/yawsi

$ yawsi --help
...
```


### Usage

I usually have these Bash functions which I combine with `fzf` and `aws` CLI tool for various use cases:

```
# AWS SSH searchable by tags
aws-ssh() {
  dns=$(yawsi --aws-profile="$1" --tags="role:$2" | fzf --exit-0 | awk '{print $7}') && ssh $dns
}

```
```
# SSH directly to an instace
sshi() {
  dns=$(yawsi --aws-profile="$1" | grep $2 | awk '{print $7}') && ssh $dns
}
```


```
# Set instance protection to an instance in an ASG
aws-set-instance-protection() {
  asg=$(yawsi --aws-profile="$1" --list-asgs | fzf --exit-0 | awk '{print $1}')
  instance_id=$(yawsi --aws-profile="$1" --asg $asg | fzf --exit-0 | awk '{print $1}')
  AWS_PROFILE=$1 aws autoscaling set-instance-protection  --instance-ids $instance_id --auto-scaling-group-name $asg --protected-from-scale-in
}
```

```
# Unset instance protection from an instance in an ASG
aws-unset-instance-protection() {
  asg=$(yawsi --aws-profile="$1" --list-asgs | fzf --exit-0 | awk '{print $1}')
  instance_id=$(yawsi --aws-profile="$1" --asg $asg | fzf --exit-0 | awk '{print $1}')
  AWS_PROFILE=$1 aws autoscaling set-instance-protection  --instance-ids $instance_id --auto-scaling-group-name $asg --no-protected-from-scale-in
}
```

```
# Show the instance protection status of instances in an ASG
aws-show-instance-protection() {
  asg=$(yawsi --aws-profile="$1" --list-asgs | fzf --exit-0 | awk '{print $1}')
  yawsi --aws-profile="$1" --asg $asg
}
```
## Building the binary

You will need `go 1.8+` installed:

```
$ BINARY_NAME=yawsi make build
```

## Building DEB packages

This needs `docker` installed and configured to be usable as 
normal user:

```
 $BINARY_NAME=yawsi DEB_PACKAGE_NAME=yawsi DEB_PACKAGE_DESCRIPTION="Yet another CLI for AWS" ./build-deb.sh
```

