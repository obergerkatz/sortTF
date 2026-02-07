# Complex module configuration with various variable types
variable "simple_string" {
  description = "A simple string variable"
  type        = string
  default     = "default-value"
}

variable "string_with_validation" {
  description = "String with complex validation"
  type        = string

  validation {
    condition     = can(regex("^[a-z0-9-]+$", var.string_with_validation))
    error_message = "Must contain only lowercase letters, numbers, and hyphens"
  }

  validation {
    condition     = length(var.string_with_validation) >= 3 && length(var.string_with_validation) <= 63
    error_message = "Length must be between 3 and 63 characters"
  }
}

variable "complex_object" {
  description = "Complex object with nested structure"
  type = object({
    name        = string
    enabled     = bool
    size        = number
    tags        = map(string)
    nested = object({
      deep_value = string
      deep_list  = list(string)
    })
  })

  default = {
    name    = "default"
    enabled = true
    size    = 10
    tags = {
      Environment = "dev"
      Project     = "example"
    }
    nested = {
      deep_value = "value"
      deep_list  = ["item1", "item2"]
    }
  }
}

variable "list_of_objects" {
  description = "List of objects"
  type = list(object({
    name = string
    cidr = string
    az   = string
  }))

  default = [
    {
      name = "subnet-1"
      cidr = "10.0.1.0/24"
      az   = "us-east-1a"
    },
    {
      name = "subnet-2"
      cidr = "10.0.2.0/24"
      az   = "us-east-1b"
    },
  ]
}

variable "map_of_objects" {
  description = "Map of objects"
  type = map(object({
    instance_type = string
    ami           = string
    disk_size     = number
  }))

  default = {
    web = {
      instance_type = "t3.small"
      ami           = "ami-12345678"
      disk_size     = 20
    }
    app = {
      instance_type = "t3.medium"
      ami           = "ami-12345678"
      disk_size     = 50
    }
  }
}

variable "optional_attributes" {
  description = "Object with optional attributes"
  type = object({
    required_field  = string
    optional_field  = optional(string, "default")
    optional_number = optional(number, 42)
    optional_bool   = optional(bool, true)
  })
}

variable "sensitive_value" {
  description = "Sensitive value"
  type        = string
  sensitive   = true
}

variable "nullable_value" {
  description = "Nullable value"
  type        = string
  default     = null
  nullable    = true
}

variable "tuple_type" {
  description = "Tuple with mixed types"
  type        = tuple([string, number, bool])
  default     = ["value", 123, true]
}

variable "set_type" {
  description = "Set of strings"
  type        = set(string)
  default     = ["value1", "value2", "value3"]
}

variable "any_type" {
  description = "Variable with any type"
  type        = any
}

locals {
  # Complex computations
  computed_map = {
    for k, v in var.map_of_objects : k => {
      instance_type = v.instance_type
      ami           = v.ami
      disk_size     = v.disk_size
      cost_estimate = v.instance_type == "t3.small" ? 10 : 20
    }
  }

  flattened_subnets = flatten([
    for item in var.list_of_objects : [
      {
        name = item.name
        cidr = item.cidr
        az   = item.az
        tags = {
          Name = item.name
          AZ   = item.az
        }
      }
    ]
  ])

  conditional_value = var.simple_string == "special" ? "special-handling" : "normal-handling"

  merged_tags = merge(
    { Environment = "production" },
    var.complex_object.tags,
    { ManagedBy = "Terraform" }
  )

  complex_for_expression = {
    for idx, item in var.list_of_objects :
    item.name => {
      index = idx
      cidr  = item.cidr
      az    = item.az
    }
  }
}

module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "~> 5.0"

  name = var.simple_string
  cidr = "10.0.0.0/16"

  azs             = ["us-east-1a", "us-east-1b", "us-east-1c"]
  private_subnets = ["10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"]
  public_subnets  = ["10.0.101.0/24", "10.0.102.0/24", "10.0.103.0/24"]

  enable_nat_gateway   = true
  single_nat_gateway   = false
  enable_dns_hostnames = true

  tags = local.merged_tags
}

module "security_group" {
  source  = "terraform-aws-modules/security-group/aws"
  version = "~> 5.0"

  for_each = var.map_of_objects

  name        = "${each.key}-sg"
  description = "Security group for ${each.key}"
  vpc_id      = module.vpc.vpc_id

  ingress_with_cidr_blocks = [
    {
      from_port   = 80
      to_port     = 80
      protocol    = "tcp"
      description = "HTTP"
      cidr_blocks = "0.0.0.0/0"
    },
  ]

  tags = local.merged_tags
}

resource "aws_instance" "this" {
  for_each = var.map_of_objects

  ami           = each.value.ami
  instance_type = each.value.instance_type

  root_block_device {
    volume_size = each.value.disk_size
    volume_type = "gp3"
  }

  vpc_security_group_ids = [module.security_group[each.key].security_group_id]
  subnet_id              = module.vpc.private_subnets[0]

  user_data = templatefile("${path.module}/user_data.sh", {
    hostname = each.key
    config   = jsonencode(var.complex_object)
  })

  tags = merge(local.merged_tags, {
    Name = each.key
    Type = each.key
  })
}

output "instance_ids" {
  description = "Map of instance IDs"
  value       = { for k, v in aws_instance.this : k => v.id }
}

output "instance_private_ips" {
  description = "Map of instance private IPs"
  value       = { for k, v in aws_instance.this : k => v.private_ip }
}

output "complex_computed_value" {
  description = "Complex computed value"
  value = {
    map    = local.computed_map
    merged = local.merged_tags
    for    = local.complex_for_expression
  }
}
