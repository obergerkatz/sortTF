variable "region" {
  default     = "us-west-2"
  description = "AWS region"
  type        = string
}

resource "aws_instance" "example" {
  ami           = "ami-123456"
  instance_type = "t3.micro"
  subnet_id     = "subnet-123456"

  tags = {
    Environment = "production"
    Name        = "example-instance"
    Project     = "test"
  }
}