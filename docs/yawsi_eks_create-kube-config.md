## yawsi eks create-kube-config

Create/update kubectl configuration

### Synopsis


Create/update kubectl configuration:

	For direct AWS users:

	    yawsi eks create-kube-config --cluster-name k8s-cluster-non-production
	
	For project teams:

	    yawsi eks create-kube-config --cluster-name k8s-cluster-non-production --project projectA --environment qa
	
	

```
yawsi eks create-kube-config [flags]
```

### Options

```
      --cluster-name string     Cluster name to create context for
      --environment string      Project environment to create context for (qa, staging, production)
  -h, --help                    help for create-kube-config
      --project string          Project name to create context for
      --show-hosts-file-entry   Show /etc/hosts file entry for private clusters (default true)
```

### SEE ALSO
* [yawsi eks](yawsi_eks.md)	 - Commands for working with AWS EKS clusters

###### Auto generated by spf13/cobra on 2-Sep-2019
