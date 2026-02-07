# Large AWS multi-region infrastructure
terraform {
  required_version = ">= 1.5"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  backend "s3" {
    bucket         = "my-terraform-state"
    key            = "multi-region/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "terraform-locks"
  }
}

provider "aws" {
  region = "us-east-1"
  alias  = "primary"

  default_tags {
    tags = local.common_tags
  }
}

provider "aws" {
  region = "us-west-2"
  alias  = "secondary"

  default_tags {
    tags = local.common_tags
  }
}

provider "aws" {
  region = "eu-west-1"
  alias  = "europe"

  default_tags {
    tags = local.common_tags
  }
}

variable "environment" {
  description = "Environment name"
  type        = string
  validation {
    condition     = contains(["dev", "staging", "prod"], var.environment)
    error_message = "Environment must be dev, staging, or prod"
  }
}

variable "project_name" {
  description = "Project name"
  type        = string
}

variable "vpc_cidrs" {
  description = "VPC CIDR blocks for each region"
  type = map(string)
  default = {
    us-east-1 = "10.0.0.0/16"
    us-west-2 = "10.1.0.0/16"
    eu-west-1 = "10.2.0.0/16"
  }
}

variable "enable_cross_region_replication" {
  description = "Enable cross-region replication for S3"
  type        = bool
  default     = true
}

variable "instance_types" {
  description = "Instance types per tier"
  type = map(string)
  default = {
    web         = "t3.small"
    app         = "t3.medium"
    database    = "db.t3.large"
    cache       = "cache.t3.micro"
  }
}

locals {
  common_tags = {
    Project     = var.project_name
    Environment = var.environment
    ManagedBy   = "Terraform"
    CostCenter  = "Engineering"
  }

  regions = ["us-east-1", "us-west-2", "eu-west-1"]

  availability_zones = {
    us-east-1 = ["us-east-1a", "us-east-1b", "us-east-1c"]
    us-west-2 = ["us-west-2a", "us-west-2b", "us-west-2c"]
    eu-west-1 = ["eu-west-1a", "eu-west-1b", "eu-west-1c"]
  }
}

# US East 1 Resources
resource "aws_vpc" "primary" {
  provider             = aws.primary
  cidr_block           = var.vpc_cidrs["us-east-1"]
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = merge(local.common_tags, {
    Name   = "${var.project_name}-vpc-us-east-1"
    Region = "us-east-1"
  })
}

resource "aws_subnet" "primary_public" {
  provider                = aws.primary
  count                   = 3
  vpc_id                  = aws_vpc.primary.id
  cidr_block              = cidrsubnet(aws_vpc.primary.cidr_block, 8, count.index)
  availability_zone       = local.availability_zones["us-east-1"][count.index]
  map_public_ip_on_launch = true

  tags = merge(local.common_tags, {
    Name = "${var.project_name}-public-${local.availability_zones["us-east-1"][count.index]}"
    Type = "public"
  })
}

resource "aws_subnet" "primary_private" {
  provider          = aws.primary
  count             = 3
  vpc_id            = aws_vpc.primary.id
  cidr_block        = cidrsubnet(aws_vpc.primary.cidr_block, 8, count.index + 10)
  availability_zone = local.availability_zones["us-east-1"][count.index]

  tags = merge(local.common_tags, {
    Name = "${var.project_name}-private-${local.availability_zones["us-east-1"][count.index]}"
    Type = "private"
  })
}

resource "aws_internet_gateway" "primary" {
  provider = aws.primary
  vpc_id   = aws_vpc.primary.id

  tags = merge(local.common_tags, {
    Name = "${var.project_name}-igw-us-east-1"
  })
}

resource "aws_eip" "primary_nat" {
  provider = aws.primary
  count    = 3
  domain   = "vpc"

  tags = merge(local.common_tags, {
    Name = "${var.project_name}-nat-eip-${count.index + 1}-us-east-1"
  })
}

resource "aws_nat_gateway" "primary" {
  provider      = aws.primary
  count         = 3
  allocation_id = aws_eip.primary_nat[count.index].id
  subnet_id     = aws_subnet.primary_public[count.index].id

  tags = merge(local.common_tags, {
    Name = "${var.project_name}-nat-${count.index + 1}-us-east-1"
  })
}

# US West 2 Resources
resource "aws_vpc" "secondary" {
  provider             = aws.secondary
  cidr_block           = var.vpc_cidrs["us-west-2"]
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = merge(local.common_tags, {
    Name   = "${var.project_name}-vpc-us-west-2"
    Region = "us-west-2"
  })
}

resource "aws_subnet" "secondary_public" {
  provider                = aws.secondary
  count                   = 3
  vpc_id                  = aws_vpc.secondary.id
  cidr_block              = cidrsubnet(aws_vpc.secondary.cidr_block, 8, count.index)
  availability_zone       = local.availability_zones["us-west-2"][count.index]
  map_public_ip_on_launch = true

  tags = merge(local.common_tags, {
    Name = "${var.project_name}-public-${local.availability_zones["us-west-2"][count.index]}"
    Type = "public"
  })
}

resource "aws_subnet" "secondary_private" {
  provider          = aws.secondary
  count             = 3
  vpc_id            = aws_vpc.secondary.id
  cidr_block        = cidrsubnet(aws_vpc.secondary.cidr_block, 8, count.index + 10)
  availability_zone = local.availability_zones["us-west-2"][count.index]

  tags = merge(local.common_tags, {
    Name = "${var.project_name}-private-${local.availability_zones["us-west-2"][count.index]}"
    Type = "private"
  })
}

resource "aws_internet_gateway" "secondary" {
  provider = aws.secondary
  vpc_id   = aws_vpc.secondary.id

  tags = merge(local.common_tags, {
    Name = "${var.project_name}-igw-us-west-2"
  })
}

# EU West 1 Resources
resource "aws_vpc" "europe" {
  provider             = aws.europe
  cidr_block           = var.vpc_cidrs["eu-west-1"]
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = merge(local.common_tags, {
    Name   = "${var.project_name}-vpc-eu-west-1"
    Region = "eu-west-1"
  })
}

resource "aws_subnet" "europe_public" {
  provider                = aws.europe
  count                   = 3
  vpc_id                  = aws_vpc.europe.id
  cidr_block              = cidrsubnet(aws_vpc.europe.cidr_block, 8, count.index)
  availability_zone       = local.availability_zones["eu-west-1"][count.index]
  map_public_ip_on_launch = true

  tags = merge(local.common_tags, {
    Name = "${var.project_name}-public-${local.availability_zones["eu-west-1"][count.index]}"
    Type = "public"
  })
}

resource "aws_subnet" "europe_private" {
  provider          = aws.europe
  count             = 3
  vpc_id            = aws_vpc.europe.id
  cidr_block        = cidrsubnet(aws_vpc.europe.cidr_block, 8, count.index + 10)
  availability_zone = local.availability_zones["eu-west-1"][count.index]

  tags = merge(local.common_tags, {
    Name = "${var.project_name}-private-${local.availability_zones["eu-west-1"][count.index]}"
    Type = "private"
  })
}

# Security Groups
resource "aws_security_group" "primary_alb" {
  provider    = aws.primary
  name        = "${var.project_name}-alb-sg-us-east-1"
  description = "Security group for ALB"
  vpc_id      = aws_vpc.primary.id

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = local.common_tags
}

resource "aws_security_group" "primary_app" {
  provider    = aws.primary
  name        = "${var.project_name}-app-sg-us-east-1"
  description = "Security group for application servers"
  vpc_id      = aws_vpc.primary.id

  ingress {
    from_port       = 8080
    to_port         = 8080
    protocol        = "tcp"
    security_groups = [aws_security_group.primary_alb.id]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = local.common_tags
}

resource "aws_security_group" "primary_db" {
  provider    = aws.primary
  name        = "${var.project_name}-db-sg-us-east-1"
  description = "Security group for database"
  vpc_id      = aws_vpc.primary.id

  ingress {
    from_port       = 5432
    to_port         = 5432
    protocol        = "tcp"
    security_groups = [aws_security_group.primary_app.id]
  }

  tags = local.common_tags
}

# S3 Buckets with Cross-Region Replication
resource "aws_s3_bucket" "primary" {
  provider = aws.primary
  bucket   = "${var.project_name}-data-us-east-1-${data.aws_caller_identity.current.account_id}"

  tags = local.common_tags
}

resource "aws_s3_bucket_versioning" "primary" {
  provider = aws.primary
  bucket   = aws_s3_bucket.primary.id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket" "secondary" {
  provider = aws.secondary
  bucket   = "${var.project_name}-data-us-west-2-${data.aws_caller_identity.current.account_id}"

  tags = local.common_tags
}

resource "aws_s3_bucket_versioning" "secondary" {
  provider = aws.secondary
  bucket   = aws_s3_bucket.secondary.id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_replication_configuration" "primary_to_secondary" {
  count    = var.enable_cross_region_replication ? 1 : 0
  provider = aws.primary
  bucket   = aws_s3_bucket.primary.id
  role     = aws_iam_role.replication[0].arn

  rule {
    id     = "replicate-all"
    status = "Enabled"

    destination {
      bucket        = aws_s3_bucket.secondary.arn
      storage_class = "STANDARD"
    }
  }

  depends_on = [
    aws_s3_bucket_versioning.primary,
    aws_s3_bucket_versioning.secondary,
  ]
}

resource "aws_iam_role" "replication" {
  count = var.enable_cross_region_replication ? 1 : 0
  name  = "${var.project_name}-s3-replication-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "s3.amazonaws.com"
        }
      }
    ]
  })

  tags = local.common_tags
}

# Route53 Health Checks and Failover
resource "aws_route53_health_check" "primary" {
  provider          = aws.primary
  fqdn              = aws_lb.primary.dns_name
  port              = 443
  type              = "HTTPS"
  resource_path     = "/health"
  failure_threshold = 3
  request_interval  = 30

  tags = merge(local.common_tags, {
    Name = "${var.project_name}-health-check-us-east-1"
  })
}

resource "aws_route53_health_check" "secondary" {
  provider          = aws.primary
  fqdn              = aws_lb.secondary.dns_name
  port              = 443
  type              = "HTTPS"
  resource_path     = "/health"
  failure_threshold = 3
  request_interval  = 30

  tags = merge(local.common_tags, {
    Name = "${var.project_name}-health-check-us-west-2"
  })
}

# Load Balancers
resource "aws_lb" "primary" {
  provider           = aws.primary
  name               = "${var.project_name}-alb-us-east-1"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.primary_alb.id]
  subnets            = aws_subnet.primary_public[*].id

  enable_deletion_protection = var.environment == "prod"

  tags = local.common_tags
}

resource "aws_lb" "secondary" {
  provider           = aws.secondary
  name               = "${var.project_name}-alb-us-west-2"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.secondary_alb.id]
  subnets            = aws_subnet.secondary_public[*].id

  enable_deletion_protection = var.environment == "prod"

  tags = local.common_tags
}

resource "aws_security_group" "secondary_alb" {
  provider    = aws.secondary
  name        = "${var.project_name}-alb-sg-us-west-2"
  description = "Security group for ALB"
  vpc_id      = aws_vpc.secondary.id

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = local.common_tags
}

# Data sources
data "aws_caller_identity" "current" {}

data "aws_availability_zones" "primary" {
  provider = aws.primary
  state    = "available"
}

data "aws_availability_zones" "secondary" {
  provider = aws.secondary
  state    = "available"
}

data "aws_availability_zones" "europe" {
  provider = aws.europe
  state    = "available"
}

# Outputs
output "primary_vpc_id" {
  description = "Primary VPC ID"
  value       = aws_vpc.primary.id
}

output "secondary_vpc_id" {
  description = "Secondary VPC ID"
  value       = aws_vpc.secondary.id
}

output "europe_vpc_id" {
  description = "Europe VPC ID"
  value       = aws_vpc.europe.id
}

output "primary_alb_dns" {
  description = "Primary ALB DNS name"
  value       = aws_lb.primary.dns_name
}

output "secondary_alb_dns" {
  description = "Secondary ALB DNS name"
  value       = aws_lb.secondary.dns_name
}

output "primary_s3_bucket" {
  description = "Primary S3 bucket name"
  value       = aws_s3_bucket.primary.id
}

output "secondary_s3_bucket" {
  description = "Secondary S3 bucket name"
  value       = aws_s3_bucket.secondary.id
}
