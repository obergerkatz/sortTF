

variable "region" {
  default     = "us-west-2"
  type        = string
  description = "AWS region"
}
resource "aws_instance" "example" {
  instance_type = "t3.micro"
  ami           = "ami-123456"
  subnet_id     = "subnet-123456"

  tags = {
    Environment = "production"
    Name        = "example-instance"
    Project     = "test"
  }
}
