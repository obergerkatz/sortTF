# Edge cases with minimal formatting
terraform { required_version = ">= 1.0" }
provider "aws" { region = "us-west-2" }
variable "region" {
  type    = string
  default = "us-west-2"
}
locals { tags = { Name = "test" } }
resource "aws_instance" "test" {
  ami           = "ami-123"
  instance_type = "t3.micro"
}