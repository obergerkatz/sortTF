resource "aws_instance" "web1" {
  ami = "ami-123"
}

resource "aws_instance" "web2" {
  ami = "ami-456"
}

resource "aws_s3_bucket" "data1" {
  bucket = "bucket1"
}

resource "aws_s3_bucket" "data2" {
  bucket = "bucket2"
}
