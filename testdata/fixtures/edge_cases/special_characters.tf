# Special characters and Unicode edge cases
variable "special_chars" {
  description = "Testing special characters: @#$%^&*()_+-=[]{}|;:',.<>?/~`"
  type        = string
  default     = "value-with-special-chars!@#$"
}

variable "unicode_chars" {
  description = "Testing Unicode: 你好世界 🌍 émojis ñ ü ö"
  type        = string
  default     = "Hello-世界-🚀"
}

variable "escaped_strings" {
  description = "Testing escaped strings: \"quotes\" and 'apostrophes' and \n newlines \t tabs"
  type        = string
  default     = "value\\with\\backslashes"
}

variable "path_separators" {
  description = "Testing path separators"
  type        = string
  default     = "C:\\Windows\\System32\\drivers\\etc\\hosts"
}

locals {
  # Map with special character keys
  special_map = {
    "key-with-dashes"      = "value1"
    "key_with_underscores" = "value2"
    "key.with.dots"        = "value3"
    "key:with:colons"      = "value4"
    "key/with/slashes"     = "value5"
    "key@with@ats"         = "value6"
  }

  # Complex regex patterns
  regex_patterns = {
    email_pattern    = "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"
    url_pattern      = "^https?://[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}(/.*)?$"
    semver_pattern   = "^(0|[1-9]\\d*)\\.(0|[1-9]\\d*)\\.(0|[1-9]\\d*)(?:-((?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\\.(?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?$"
    ipv4_pattern     = "^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$"
  }

  # JSON with special characters
  json_config = jsonencode({
    "api-key" = "secret-key-123!@#"
    "url"     = "https://example.com/path?query=value&other=123"
    "config"  = {
      "enable_feature_A" = true
      "enable_feature_B" = false
    }
  })
}

resource "null_resource" "heredoc_with_special_chars" {
  provisioner "local-exec" {
    command = <<-EOT
      #!/bin/bash
      echo "Testing special chars: $VAR, ${VAR}, @#$%"
      echo 'Single quotes with $VAR'
      echo "Double quotes with ${VAR}"
      cat <<-EOF
        Nested heredoc!
        With special chars: @#$%^&*()
      EOF
    EOT
  }
}

# Resource names with special patterns
resource "aws_s3_bucket" "bucket-with-dashes" {
  bucket = "my-bucket-${var.environment}-2024"
}

resource "aws_s3_bucket" "bucket_with_underscores" {
  bucket = "my_bucket_${var.environment}_2024"
}
