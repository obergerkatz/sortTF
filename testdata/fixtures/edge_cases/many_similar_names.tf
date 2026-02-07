# Many resources with similar names - tests alphabetical sorting
resource "aws_subnet" "private_subnet_1a" {
  cidr_block = "10.0.1.0/24"
}

resource "aws_subnet" "private_subnet_1b" {
  cidr_block = "10.0.2.0/24"
}

resource "aws_subnet" "private_subnet_1c" {
  cidr_block = "10.0.3.0/24"
}

resource "aws_subnet" "public_subnet_1a" {
  cidr_block = "10.0.101.0/24"
}

resource "aws_subnet" "public_subnet_1b" {
  cidr_block = "10.0.102.0/24"
}

resource "aws_subnet" "public_subnet_1c" {
  cidr_block = "10.0.103.0/24"
}

# Test that resources with numbers sort correctly
resource "aws_instance" "web_1" {
  ami = "ami-12345678"
}

resource "aws_instance" "web_2" {
  ami = "ami-12345678"
}

resource "aws_instance" "web_10" {
  ami = "ami-12345678"
}

resource "aws_instance" "web_20" {
  ami = "ami-12345678"
}

resource "aws_instance" "web_100" {
  ami = "ami-12345678"
}

# Test resources with prefixes
resource "aws_security_group" "app_backend" {
  name = "backend"
}

resource "aws_security_group" "app_cache" {
  name = "cache"
}

resource "aws_security_group" "app_database" {
  name = "database"
}

resource "aws_security_group" "app_frontend" {
  name = "frontend"
}

resource "aws_security_group" "app_lb" {
  name = "loadbalancer"
}

# Test data sources with similar names
data "aws_availability_zones" "available" {
  state = "available"
}

data "aws_availability_zones" "available_excluding_local" {
  state = "available"
  filter {
    name   = "zone-type"
    values = ["availability-zone"]
  }
}

# Test variables with similar names
variable "instance_type_app" {
  type = string
}

variable "instance_type_cache" {
  type = string
}

variable "instance_type_database" {
  type = string
}

variable "instance_type_web" {
  type = string
}

# Test locals with similar names
locals {
  subnet_id_1a = aws_subnet.public_subnet_1a.id
  subnet_id_1b = aws_subnet.public_subnet_1b.id
  subnet_id_1c = aws_subnet.public_subnet_1c.id
  subnet_ids   = [local.subnet_id_1a, local.subnet_id_1b, local.subnet_id_1c]
}

# Test modules with similar names
module "app_backend" {
  source = "./modules/app"
  name   = "backend"
}

module "app_frontend" {
  source = "./modules/app"
  name   = "frontend"
}

module "app_worker" {
  source = "./modules/app"
  name   = "worker"
}

# Test outputs with similar names
output "backend_instance_id" {
  value = module.app_backend.instance_id
}

output "frontend_instance_id" {
  value = module.app_frontend.instance_id
}

output "worker_instance_id" {
  value = module.app_worker.instance_id
}
