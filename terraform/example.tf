provider "aws" {}

resource "aws_instance" "example" {
  ami           = "ami-2757f631"
  instance_type = "t2.micro"
}

data "aws_availability_zones" "available" {}

resource "aws_vpc" "example" {
  cidr_block = "10.0.0.0/16"
  tags = "${
    map(
      "Name", "terraform-eks-node",
      "kubernetes.io/cluster/terraform-eks", "shared",
    )
  }"
}

resource "aws_subnet" "example" {
  count = 2
  availability_zone = "${data.aws_availability_zones.available.names[count.index]}"
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = "${aws_vpc.example.id}"
  tags = "${
    map(
     "Name", "terraform-eks-node",
     "kubernetes.io/cluster/terraform-eks", "shared",
    )
  }"
}

resource "aws_internet_gateway" "example" {
  vpc_id = "${aws_vpc.example.id}"

  tags = {
    Name = "terraform-eks"
  }
}

resource "aws_route_table" "example" {
  vpc_id = "${aws_vpc.example.id}"

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = "${aws_internet_gateway.example.id}"
  }
}

resource "aws_route_table_association" "example" {
  count = 2

  subnet_id      = "${aws_subnet.example.*.id[count.index]}"
  route_table_id = "${aws_route_table.example.id}"
}

resource "aws_security_group" "tf-eks-master" {
  name        = "terraform-eks-cluster"
  description = "Cluster communication with worker nodes"
  vpc_id      = "${aws_vpc.example.id}"

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
      Name = "terraform-eks"
  }
}

resource "aws_security_group" "tf-eks-node" {
    name        = "terraform-eks-node"
    description = "Security group for all nodes in the cluster"
    vpc_id      = "${aws_vpc.example.id}"
 
    egress {
        from_port   = 0
        to_port     = 0
        protocol    = "-1"
        cidr_blocks = ["0.0.0.0/0"]
    }
 
    tags = "${
      map(
        "Name", "terraform-eks-node",
        "kubernetes.io/cluster/terraform-eks", "owned",
      )
    }"
}

resource "aws_iam_role" "example-cluster" {
  name = "example-kubectl-access-role"

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "eks.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "example-cluster-AmazonEKSClusterPolicy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSClusterPolicy"
  role       = "${aws_iam_role.example-cluster.name}"
}

resource "aws_iam_role_policy_attachment" "example-cluster-AmazonEKSServicePolicy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSServicePolicy"
  role       = "${aws_iam_role.example-cluster.name}"
}

resource "aws_eks_cluster" "tf_eks" {
  name            = "terraform-eks"
  role_arn        = "${aws_iam_role.example-cluster.arn}"

  vpc_config {
    security_group_ids = ["${aws_security_group.tf-eks-master.id}"]
    subnet_ids         = "${aws_subnet.example.*.id}"
  }

  depends_on = [
    "aws_iam_role_policy_attachment.example-cluster-AmazonEKSClusterPolicy",
    "aws_iam_role_policy_attachment.example-cluster-AmazonEKSServicePolicy",
  ]
}

resource "aws_iam_role" "tf-eks-node" {
  name = "terraform-eks-node"
 
  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "tf-eks-node-AmazonEKSWorkerNodePolicy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy"
  role       = "${aws_iam_role.tf-eks-node.name}"
}
 
resource "aws_iam_role_policy_attachment" "tf-eks-node-AmazonEKS_CNI_Policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy"
  role       = "${aws_iam_role.tf-eks-node.name}"
}
 
resource "aws_iam_role_policy_attachment" "tf-eks-node-AmazonEC2ContainerRegistryReadOnly" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
  role       = "${aws_iam_role.tf-eks-node.name}"
}
 
resource "aws_iam_instance_profile" "node" {
  name = "terraform-eks-node"
  role = "${aws_iam_role.tf-eks-node.name}"
}

data "aws_ami" "eks-worker" {
  filter {
    name   = "name"
    values = ["amazon-eks-node-${aws_eks_cluster.tf_eks.version}-v*"]
  }
 
  most_recent = true
  owners      = ["310254572914"] # Amazon EKS AMI Account ID
}

data "aws_region" "current" {}

locals {
  tf-eks-node-userdata = <<USERDATA
#!/bin/bash
set -o trace
/etc/eks/bootstrap.sh --apiserver-endpoint '${aws_eks_cluster.tf_eks.endpoint}' --b64-cluster-ca '${aws_eks_cluster.tf_eks.certificate_authority.0.data}' 'terraform-eks'
USERDATA
}

resource "aws_launch_configuration" "example" {
  associate_public_ip_address = true
  iam_instance_profile        = "${aws_iam_instance_profile.node.name}"
  image_id                    = "${data.aws_ami.eks-worker.id}"
  instance_type               = "m4.large"
  name_prefix                 = "terraform-eks"
  security_groups             = ["${aws_security_group.tf-eks-node.id}"]
  user_data_base64            = "${base64encode(local.tf-eks-node-userdata)}"

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_autoscaling_group" "example" {
  desired_capacity     = 2
  launch_configuration = "${aws_launch_configuration.example.id}"
  max_size             = 2
  min_size             = 1
  name                 = "terraform-eks"
  vpc_zone_identifier  = "${aws_subnet.example.*.id}"

  tag {
    key                 = "Name"
    value               = "terraform-eks"
    propagate_at_launch = true
  }

  tag {
    key                 = "kubernetes.io/cluster/terraform-eks"
    value               = "owned"
    propagate_at_launch = true
  }
}

data "aws_eks_cluster_auth" "cluster_auth" {
  name = "terraform-eks"
}

provider "kubernetes" {
  host                   = "${aws_eks_cluster.tf_eks.endpoint}"
  cluster_ca_certificate = "${base64decode(aws_eks_cluster.tf_eks.certificate_authority.0.data)}"
  token                  = "${data.aws_eks_cluster_auth.cluster_auth.token}"
  load_config_file       = false
}

resource "kubernetes_config_map" "aws_auth_configmap" {
  metadata {
    name      = "aws-auth"
    namespace = "kube-system"
  }

  data = {
    mapUsers = <<YAML
- userarn: arn:aws:iam::310254572914:root
  username: Administrator
YAML
    mapRoles = <<YAML
- rolearn: ${aws_iam_role.tf-eks-node.arn}
  username: system:node:{{EC2PrivateDNSName}}
  groups:
    - system:bootstrappers
    - system:nodes
- rolearn: ${aws_iam_role.example-cluster.arn}
  username: kubectl-access-user
  groups:
    - system:masters
YAML
  }
}
