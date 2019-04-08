
# This is a generated file, do not hand edit. See README at the
# root of the repository

resource "aws_network_acl_rule" "rule_SubnetA_ingress_101" {

    network_acl_id = "${lookup(local.network_acl_ids_map, "SubnetA")}"
    egress = false
    rule_number = 101
    rule_action = "allow"
    cidr_block = "127.0.0.1/32"
    protocol = "tcp"
    from_port = 22
    to_port = 30


}
resource "aws_network_acl_rule" "rule_SubnetA_ingress_102" {

    network_acl_id = "${lookup(local.network_acl_ids_map, "SubnetA")}"
    egress = false
    rule_number = 102
    rule_action = "allow"
    cidr_block = "127.0.0.1/32"
    protocol = "tcp"
    from_port = 22
    to_port = 30


}

