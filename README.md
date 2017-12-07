## Yet another AWS CLI

### Install

```
$ go get github.com/amitsaha/yawsi

$ yawsi --help
...
```


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
# SSH directly to an instace
sshi() {
  dns=$(yawsi --aws-profile="$1" | grep $2 | awk '{print $7}') && ssh $dns
}
```

```
aws-set-instance-protection() {
  asg=$(yawsi --aws-profile="$1" --list-asgs | fzf --exit-0 | awk '{print $1}')
  instance_id=$(yawsi --aws-profile="$1" --asg $asg | fzf --exit-0 | awk '{print $1}')
  AWS_PROFILE=$1 aws autoscaling set-instance-protection  --instance-ids $instance_id --auto-scaling-group-name $asg --protected-from-scale-in
}
```

```
aws-unset-instance-protection() {
  asg=$(yawsi --aws-profile="$1" --list-asgs | fzf --exit-0 | awk '{print $1}')
  instance_id=$(yawsi --aws-profile="$1" --asg $asg | fzf --exit-0 | awk '{print $1}')
  AWS_PROFILE=$1 aws autoscaling set-instance-protection  --instance-ids $instance_id --auto-scaling-group-name $asg --no-protected-from-scale-in
}
```

```
aws-show-instance-protection() {
  asg=$(yawsi --aws-profile="$1" --list-asgs | fzf --exit-0 | awk '{print $1}')
  yawsi --aws-profile="$1" --asg $asg
}
```

