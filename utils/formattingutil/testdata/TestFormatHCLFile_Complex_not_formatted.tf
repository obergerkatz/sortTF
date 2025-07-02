# Complex Terraform configuration with poor formatting
terraform{
  required_version=">= 1.0"
  backend"s3"{
    bucket="terraform-state-bucket"
    key="prod/terraform.tfstate"
    region="us-west-2"
    encrypt=true
  }
}
locals{
  common_tags={
    Owner="DevOps Team"
    Environment="production"
    CostCenter="IT"
  }
}
#!/bin/bash