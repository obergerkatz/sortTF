#!/bin/bash
# Generate comprehensive test fixtures for sortTF

set -e

FIXTURES_DIR="$(cd "$(dirname "$0")" && pwd)"

# Syntax Edge Cases
cat > "$FIXTURES_DIR/syntax/whitespace.tf" << 'EOF'




EOF

cat > "$FIXTURES_DIR/syntax/comments_hash.tf" << 'EOF'
# This is a hash comment
variable "test" {
  type = string # inline comment
}
# Another comment
EOF

cat > "$FIXTURES_DIR/syntax/comments_double_slash.tf" << 'EOF'
// This is a double slash comment
variable "test" {
  type = string // inline comment
}
// Another comment
EOF

cat > "$FIXTURES_DIR/syntax/comments_multiline.tf" << 'EOF'
/* This is a
   multiline
   block comment */
variable "test" {
  type = string /* inline block comment */
}
EOF

cat > "$FIXTURES_DIR/syntax/comments_mixed.tf" << 'EOF'
# Hash comment
// Double slash comment
/* Block comment */
variable "test" {
  type = string
  # Hash
  // Slash
  /* Block */
}
EOF

cat > "$FIXTURES_DIR/syntax/heredoc_standard.tf" << 'EOF'
locals {
  user_data = <<EOF
#!/bin/bash
echo "Hello World"
EOF
}
EOF

cat > "$FIXTURES_DIR/syntax/heredoc_indented.tf" << 'EOF'
locals {
  user_data = <<-EOF
    #!/bin/bash
    echo "Indented"
  EOF
}
EOF

cat > "$FIXTURES_DIR/syntax/multiline_strings.tf" << 'EOF'
variable "description" {
  default = "This is a
multiline
string"
}
EOF

cat > "$FIXTURES_DIR/syntax/unclosed_brace.tf" << 'EOF'
resource "aws_instance" "web" {
  ami = "ami-12345"
EOF

cat > "$FIXTURES_DIR/syntax/unclosed_quote.tf" << 'EOF'
variable "test {
  type = string
}
EOF

# Structure Edge Cases
cat > "$FIXTURES_DIR/structure/all_block_types.tf" << 'EOF'
terraform {
  required_version = ">= 1.0"
}

provider "aws" {
  region = "us-west-2"
}

variable "name" {
  type = string
}

locals {
  tags = {}
}

data "aws_ami" "ubuntu" {
  most_recent = true
}

resource "aws_instance" "web" {
  ami = "ami-123"
}

module "vpc" {
  source = "./vpc"
}

output "id" {
  value = aws_instance.web.id
}
EOF

cat > "$FIXTURES_DIR/structure/nested_blocks.tf" << 'EOF'
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
  }

  backend "s3" {
    bucket = "terraform-state"
    key    = "state"
    region = "us-west-2"
  }
}
EOF

cat > "$FIXTURES_DIR/structure/dynamic_blocks.tf" << 'EOF'
resource "aws_security_group" "app" {
  name = "app"

  dynamic "ingress" {
    for_each = var.ingress_rules
    content {
      from_port   = ingress.value.from_port
      to_port     = ingress.value.to_port
      protocol    = ingress.value.protocol
      cidr_blocks = ingress.value.cidr_blocks
    }
  }
}
EOF

cat > "$FIXTURES_DIR/structure/repeated_blocks.tf" << 'EOF'
resource "aws_instance" "web1" {
  ami = "ami-123"
}

resource "aws_instance" "web2" {
  ami = "ami-456"
}

resource "aws_s3_bucket" "data1" {
  bucket = "bucket1"
}

resource "aws_s3_bucket" "data2" {
  bucket = "bucket2"
}
EOF

# Type/Value Edge Cases
cat > "$FIXTURES_DIR/types/nulls.tf" << 'EOF'
variable "nullable" {
  type    = string
  default = null
}

resource "aws_instance" "web" {
  ami                    = "ami-123"
  user_data              = null
  iam_instance_profile   = null
}
EOF

cat > "$FIXTURES_DIR/types/empty_collections.tf" << 'EOF'
variable "empty_list" {
  default = []
}

variable "empty_map" {
  default = {}
}

resource "aws_instance" "web" {
  ami           = "ami-123"
  tags          = {}
  security_groups = []
}
EOF

cat > "$FIXTURES_DIR/types/nested_collections.tf" << 'EOF'
variable "complex" {
  type = map(object({
    name = string
    ports = list(number)
    config = map(string)
  }))
  default = {
    app = {
      name = "web"
      ports = [80, 443]
      config = {
        env = "prod"
      }
    }
  }
}
EOF

cat > "$FIXTURES_DIR/types/numbers.tf" << 'EOF'
variable "int" {
  default = 42
}

variable "float" {
  default = 3.14
}

variable "scientific" {
  default = 1.23e-4
}

variable "negative" {
  default = -100
}
EOF

cat > "$FIXTURES_DIR/types/strings_as_numbers.tf" << 'EOF'
variable "string_number" {
  default = "123"
}

variable "actual_number" {
  default = 123
}

resource "aws_instance" "web" {
  ami = "ami-123"
  count = var.actual_number
}
EOF

cat > "$FIXTURES_DIR/types/booleans.tf" << 'EOF'
variable "enabled_true" {
  default = true
}

variable "enabled_false" {
  default = false
}

resource "aws_instance" "web" {
  ami                         = "ami-123"
  associate_public_ip_address = true
  monitoring                  = false
}
EOF

# Control Flow
cat > "$FIXTURES_DIR/control/for_each.tf" << 'EOF'
resource "aws_instance" "web" {
  for_each      = var.instances
  ami           = "ami-123"
  instance_type = each.value.type
  tags = {
    Name = each.key
  }
}
EOF

cat > "$FIXTURES_DIR/control/count.tf" << 'EOF'
resource "aws_instance" "web" {
  count         = var.instance_count
  ami           = "ami-123"
  instance_type = "t2.micro"
  tags = {
    Name = "web-${count.index}"
  }
}
EOF

cat > "$FIXTURES_DIR/control/conditionals.tf" << 'EOF'
resource "aws_instance" "web" {
  count         = var.enabled ? 1 : 0
  ami           = "ami-123"
  instance_type = var.instance_type != "" ? var.instance_type : "t2.micro"
}
EOF

cat > "$FIXTURES_DIR/control/interpolation.tf" << 'EOF'
variable "env" {
  default = "prod"
}

resource "aws_instance" "web" {
  ami = "ami-123"
  tags = {
    Name        = "web-${var.env}"
    Environment = var.env
  }
}
EOF

cat > "$FIXTURES_DIR/control/functions.tf" << 'EOF'
locals {
  tags = merge(
    {
      Environment = var.env
    },
    var.common_tags
  )

  cidr_blocks = [
    for i in range(3) :
    "10.0.${i}.0/24"
  ]

  uppercase_name = upper(var.name)
  joined_names   = join(",", var.names)
}
EOF

# Realistic Scenarios
cat > "$FIXTURES_DIR/realistic/aws_infrastructure.tf" << 'EOF'
terraform {
  required_version = ">= 1.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
  }
}

provider "aws" {
  region = var.region
}

variable "region" {
  description = "AWS region"
  type        = string
  default     = "us-west-2"
}

variable "environment" {
  description = "Environment name"
  type        = string
}

locals {
  common_tags = {
    Environment = var.environment
    ManagedBy   = "Terraform"
    Project     = "MyApp"
  }
}

data "aws_ami" "ubuntu" {
  most_recent = true
  owners      = ["099720109477"]

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-server-*"]
  }
}

resource "aws_vpc" "main" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true
  tags                 = merge(local.common_tags, { Name = "main-vpc" })
}

resource "aws_subnet" "public" {
  vpc_id                  = aws_vpc.main.id
  cidr_block              = "10.0.1.0/24"
  availability_zone       = "${var.region}a"
  map_public_ip_on_launch = true
  tags                    = merge(local.common_tags, { Name = "public-subnet" })
}

resource "aws_security_group" "web" {
  name        = "web-sg"
  description = "Security group for web servers"
  vpc_id      = aws_vpc.main.id

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

  tags = merge(local.common_tags, { Name = "web-sg" })
}

resource "aws_instance" "web" {
  ami                    = data.aws_ami.ubuntu.id
  instance_type          = "t3.micro"
  subnet_id              = aws_subnet.public.id
  vpc_security_group_ids = [aws_security_group.web.id]

  user_data = <<-EOF
    #!/bin/bash
    apt-get update
    apt-get install -y nginx
    systemctl start nginx
  EOF

  tags = merge(local.common_tags, { Name = "web-server" })
}

output "vpc_id" {
  description = "VPC ID"
  value       = aws_vpc.main.id
}

output "instance_id" {
  description = "Web server instance ID"
  value       = aws_instance.web.id
}

output "instance_public_ip" {
  description = "Web server public IP"
  value       = aws_instance.web.public_ip
}
EOF

cat > "$FIXTURES_DIR/realistic/terragrunt.hcl" << 'EOF'
include "root" {
  path = find_in_parent_folders()
}

terraform {
  source = "git::https://github.com/terraform-aws-modules/terraform-aws-vpc.git?ref=v3.0.0"
}

inputs = {
  name = "my-vpc"
  cidr = "10.0.0.0/16"

  azs             = ["us-west-2a", "us-west-2b", "us-west-2c"]
  private_subnets = ["10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"]
  public_subnets  = ["10.0.101.0/24", "10.0.102.0/24", "10.0.103.0/24"]

  enable_nat_gateway = true
  enable_vpn_gateway = false

  tags = {
    Environment = "prod"
    Terraform   = "true"
  }
}
EOF

echo "✅ Fixture generation complete!"
echo "Created fixtures in: $FIXTURES_DIR"
find "$FIXTURES_DIR" -type f -name "*.tf" -o -name "*.hcl" | wc -l | xargs echo "Total fixture files:"
