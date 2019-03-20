resource "aws_vpc" "vpc1" {
  cidr_block       = "10.0.0.0/16"
}

resource "aws_vpc" "vpc2" {
    cidr_block = "10.1.0.0/16"  
}

resource "aws_subnet" "vpc1_subnet1" {
  vpc_id     = "${aws_vpc.vpc1.id}"
  cidr_block = "10.0.1.0/24"

  tags = {
    Name = "vpc1_subnet1"
  }
}

resource "aws_subnet" "vpc2_subnet1" {
  vpc_id     = "${aws_vpc.vpc2.id}"
  cidr_block = "10.1.1.0/24"

  tags = {
    Name = "vpc1_subnet1"
  }
}
