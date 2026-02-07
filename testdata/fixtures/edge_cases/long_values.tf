# Long attribute values and complex expressions
variable "very_long_description" {
  description = "This is a very long description that spans multiple lines and contains a lot of text to test how the tool handles very long string values that might need special formatting or handling in the output"
  type        = string
  default     = "This is an extremely long default value with lots and lots of text that goes on and on and on to test how the sorting tool handles very long attribute values that might cause line wrapping or other formatting challenges"
}

variable "complex_validation" {
  type = string
  validation {
    condition     = can(regex("^(prod|staging|dev|test)-(us|eu|asia)-(east|west|central)-[0-9]{1,3}$", var.complex_validation))
    error_message = "Value must match the pattern: {environment}-{region}-{location}-{number} where environment is prod/staging/dev/test, region is us/eu/asia, location is east/west/central, and number is 1-999."
  }
}

locals {
  # Very long expression
  ultra_long_computed_value = format(
    "%s-%s-%s-%s-%s-%s-%s-%s-%s-%s",
    var.environment,
    var.region,
    var.availability_zone,
    var.application_name,
    var.component_name,
    var.team_name,
    var.cost_center,
    var.project_code,
    random_id.suffix.hex,
    formatdate("YYYY-MM-DD", timestamp())
  )

  # Complex nested map with long keys
  configuration_with_very_long_keys = {
    "application_configuration_section_with_database_settings" = {
      "primary_database_connection_string_with_ssl_enabled"   = "postgresql://user:pass@db.example.com:5432/dbname?sslmode=require&connection_timeout=30"
      "secondary_database_connection_string_with_ssl_enabled" = "postgresql://user:pass@db-replica.example.com:5432/dbname?sslmode=require&connection_timeout=30"
    }
    "application_configuration_section_with_cache_settings" = {
      "primary_redis_connection_string_with_cluster_mode"   = "rediss://default:password@redis-cluster-0001-001.example.com:6379/0?ssl=true&cluster_mode=enabled"
      "secondary_redis_connection_string_with_cluster_mode" = "rediss://default:password@redis-cluster-0002-001.example.com:6379/0?ssl=true&cluster_mode=enabled"
    }
  }

  # Long list
  all_supported_aws_regions = [
    "us-east-1", "us-east-2", "us-west-1", "us-west-2",
    "ca-central-1",
    "eu-west-1", "eu-west-2", "eu-west-3", "eu-central-1", "eu-north-1",
    "ap-south-1", "ap-northeast-1", "ap-northeast-2", "ap-northeast-3",
    "ap-southeast-1", "ap-southeast-2",
    "sa-east-1",
    "af-south-1",
    "me-south-1",
  ]
}

resource "aws_iam_policy" "complex_policy" {
  name        = "complex-policy-with-many-permissions"
  description = "A complex IAM policy with many permissions for testing"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:PutObject",
          "s3:DeleteObject",
          "s3:ListBucket",
          "s3:GetBucketLocation",
          "s3:GetBucketVersioning",
          "s3:PutLifecycleConfiguration",
          "s3:GetLifecycleConfiguration",
        ]
        Resource = [
          "arn:aws:s3:::my-application-bucket-${var.environment}-${var.region}-${random_id.suffix.hex}",
          "arn:aws:s3:::my-application-bucket-${var.environment}-${var.region}-${random_id.suffix.hex}/*",
        ]
      },
      {
        Effect = "Allow"
        Action = [
          "dynamodb:GetItem",
          "dynamodb:PutItem",
          "dynamodb:UpdateItem",
          "dynamodb:DeleteItem",
          "dynamodb:Query",
          "dynamodb:Scan",
          "dynamodb:BatchGetItem",
          "dynamodb:BatchWriteItem",
        ]
        Resource = "arn:aws:dynamodb:${var.region}:${data.aws_caller_identity.current.account_id}:table/my-application-table-${var.environment}"
      },
    ]
  })
}
