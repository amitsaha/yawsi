
# This is a generated file, do not hand edit. See README at the
# root of the repository

resource "aws_network_acl_rule" "rule_SubnetA_ingress_" {

    network_acl_id = &#34;&#34;
    egress = false
    rule_number = 
    rule_action = &#34;allow&#34;
    cidr_block = &#34;127.0.0.1/32&#34;
    protocol = &#34;tcp&#34;
    from_port = 22
    to_port = 30


}
resource "aws_network_acl_rule" "rule_SubnetA_egress_" {

    network_acl_id = &#34;&#34;
    egress = true
    rule_number = 
    rule_action = &#34;allow&#34;
    cidr_block = &#34;127.0.0.1/32&#34;
    protocol = &#34;tcp&#34;
    from_port = 22
    to_port = 30


}

