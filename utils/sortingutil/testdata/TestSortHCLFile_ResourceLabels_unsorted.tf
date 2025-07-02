


resource "aws_instance" "app" {
  ami           = "ami-123456"
  instance_type = "t3.small"
}
resource "aws_instance" "web1" {
  ami           = "ami-123456"
  instance_type = "t3.micro"
}
resource "aws_instance" "web2" {
  ami           = "ami-123456"
  instance_type = "t3.micro"
}





resource "aws_instance" "web3" {
  ami           = "ami-123456"
  instance_type = "t3.micro"
}
resource "aws_s3_bucket" "data" {
  bucket = "my-data-bucket"
}


resource "aws_s3_bucket" "logs" {
  bucket = "my-logs-bucket"
}
