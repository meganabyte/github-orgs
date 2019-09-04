provider "aws" {
  region = "us-east-1"
}

resource "aws_instance" "example" {
  ami           = "ami-2757f631"
  instance_type = "t2.micro"
}

data "aws_availability_zones" "available" {}

resource "aws_vpc" "example" {
  cidr_block = "10.0.0.0/16"
  tags = "${
    map(
      "Name", "meg-app-node",
      "kubernetes.io/cluster/meg-app", "shared",
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
     "Name", "meg-app-node",
     "kubernetes.io/cluster/meg-app", "shared",
    )
  }"
}

resource "aws_internet_gateway" "example" {
  vpc_id = "${aws_vpc.example.id}"

  tags = {
    Name = "meg-app"
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
  name        = "meg-app-cluster"
  description = "Cluster communication with worker nodes"
  vpc_id      = "${aws_vpc.example.id}"

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
      Name = "meg-app"
  }
}

resource "aws_security_group_rule" "tf-eks-master-ingress-workstation-https" {
  cidr_blocks       = ["50.230.15.130/32"]
  description       = "Allow workstation to communicate with the cluster API Server"
  from_port         = 443
  protocol          = "tcp"
  security_group_id = "${aws_security_group.tf-eks-master.id}"
  to_port           = 443
  type              = "ingress"
}

resource "aws_security_group" "tf-eks-node" {
    name        = "meg-app-node"
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
        "Name", "meg-app-node",
        "kubernetes.io/cluster/meg-app", "owned",
      )
    }"
}

resource "aws_security_group_rule" "tf-eks-node-ingress-self" {
  description              = "Allow node to communicate with each other"
  from_port                = 0
  protocol                 = "-1"
  security_group_id        = "${aws_security_group.tf-eks-node.id}"
  source_security_group_id = "${aws_security_group.tf-eks-node.id}"
  to_port                  = 65535
  type                     = "ingress"
}

resource "aws_security_group_rule" "tf-eks-node-ingress-cluster" {
  description              = "Allow worker Kubelets and pods to receive communication from the cluster control plane"
  from_port                = 1025
  protocol                 = "tcp"
  security_group_id        = "${aws_security_group.tf-eks-node.id}"
  source_security_group_id = "${aws_security_group.tf-eks-master.id}"
  to_port                  = 65535
  type                     = "ingress"
}

resource "aws_security_group_rule" "tf-eks-node-ingress-node-https" {
  description              = "Allow pods to communicate with the cluster API Server"
  from_port                = 443
  protocol                 = "tcp"
  security_group_id        = "${aws_security_group.tf-eks-master.id}"
  source_security_group_id = "${aws_security_group.tf-eks-node.id}"
  to_port                  = 443
  type                     = "ingress"
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
  name            = "meg-app"
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
  name = "meg-app-node"
 
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
  name = "meg-app-node"
  role = "${aws_iam_role.tf-eks-node.name}"
}

data "aws_ami" "eks-worker" {
  filter {
    name   = "name"
    values = ["amazon-eks-node-${aws_eks_cluster.tf_eks.version}-v*"]
  }
 
  most_recent = true
  owners      = ["amazon"] 
}

data "aws_region" "current" {}

locals {
  tf-eks-node-userdata = <<USERDATA
#!/bin/bash
set -o trace
/etc/eks/bootstrap.sh --apiserver-endpoint '${aws_eks_cluster.tf_eks.endpoint}' --b64-cluster-ca '${aws_eks_cluster.tf_eks.certificate_authority.0.data}' 'meg-app'
USERDATA
}

resource "aws_launch_configuration" "example" {
  associate_public_ip_address = true
  iam_instance_profile        = "${aws_iam_instance_profile.node.name}"
  image_id                    = "${data.aws_ami.eks-worker.id}"
  instance_type               = "t2.micro"
  name_prefix                 = "meg-app"
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
  name                 = "meg-app"
  vpc_zone_identifier  = "${aws_subnet.example.*.id}"

  tag {
    key                 = "Name"
    value               = "meg-app"
    propagate_at_launch = true
  }

  tag {
    key                 = "kubernetes.io/cluster/meg-app"
    value               = "owned"
    propagate_at_launch = true
  }
}

data "aws_eks_cluster_auth" "cluster_auth" {
  name = "meg-app"
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
