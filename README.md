## Yet another AWS CLI

[] Sub commands
[] AWS: launch more like this

```
# AWS SSH searchable by tags
aws-ssh() {
  dns=$(cloudi --aws-profile="$1" --tags="role:$2" | fzf --exit-0 | awk '{print $7}') && ssh $dns
}

```
```
# SSH directly to an instace
sshi() {
  dns=$(cloudi --aws-profile="$1" | grep $2 | awk '{print $7}') && ssh $dns
}
```
