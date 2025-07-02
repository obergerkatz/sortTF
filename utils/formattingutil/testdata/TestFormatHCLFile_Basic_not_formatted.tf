# This is a poorly formatted Terraform file
terraform{
  required_version=">= 1.0"
  backend"s3"{
    bucket="terraform-state-bucket"
    key="prod/terraform.tfstate"
  }
}
locals{
  common_tags={
    Owner="DevOps Team"
    Environment="production"
  }
}