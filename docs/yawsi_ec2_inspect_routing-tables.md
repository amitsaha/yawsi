## yawsi ec2 inspect routing-tables

Routing table entries associated with an instance

### Synopsis


Show the routing table associated with an EC2 instance:

	$ yawsi ec2  inspect routing-tables i-06d80024e0df241da

	RoutTableID     Main    Destination     Target
	-----------     ----    ------------    ------
	rtb-d1df42b5    false   172.31.0.0/16   local
	rtb-d24342b5    false   pl-6ca54005(S3) vpce-6b2ecf02
	rtb-942315f1    true    172.31.0.0/16   pcx-cd9541a4 (vpc-20988a4 - VPCA)
	rtb-63caa9f1    true    0.0.0.0/0       igw-121234
	

```
yawsi ec2 inspect routing-tables [flags]
```

### Options

```
  -h, --help   help for routing-tables
```

### SEE ALSO
* [yawsi ec2 inspect](yawsi_ec2_inspect.md)	 - Perform various checks

###### Auto generated by spf13/cobra on 21-Mar-2019