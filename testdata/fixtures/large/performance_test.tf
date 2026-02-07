# Very large file for performance testing - 1000+ lines
# This file tests sorting performance on large configurations

terraform {
  required_version = ">= 1.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = "us-east-1"
}

# Generate 100 variables
variable "var_001" { type = string }
variable "var_002" { type = string }
variable "var_003" { type = string }
variable "var_004" { type = string }
variable "var_005" { type = string }
variable "var_006" { type = string }
variable "var_007" { type = string }
variable "var_008" { type = string }
variable "var_009" { type = string }
variable "var_010" { type = string }
variable "var_011" { type = string }
variable "var_012" { type = string }
variable "var_013" { type = string }
variable "var_014" { type = string }
variable "var_015" { type = string }
variable "var_016" { type = string }
variable "var_017" { type = string }
variable "var_018" { type = string }
variable "var_019" { type = string }
variable "var_020" { type = string }
variable "var_021" { type = string }
variable "var_022" { type = string }
variable "var_023" { type = string }
variable "var_024" { type = string }
variable "var_025" { type = string }
variable "var_026" { type = string }
variable "var_027" { type = string }
variable "var_028" { type = string }
variable "var_029" { type = string }
variable "var_030" { type = string }
variable "var_031" { type = string }
variable "var_032" { type = string }
variable "var_033" { type = string }
variable "var_034" { type = string }
variable "var_035" { type = string }
variable "var_036" { type = string }
variable "var_037" { type = string }
variable "var_038" { type = string }
variable "var_039" { type = string }
variable "var_040" { type = string }
variable "var_041" { type = string }
variable "var_042" { type = string }
variable "var_043" { type = string }
variable "var_044" { type = string }
variable "var_045" { type = string }
variable "var_046" { type = string }
variable "var_047" { type = string }
variable "var_048" { type = string }
variable "var_049" { type = string }
variable "var_050" { type = string }

# Generate locals
locals {
  local_001 = "value_001"
  local_002 = "value_002"
  local_003 = "value_003"
  local_004 = "value_004"
  local_005 = "value_005"
  local_006 = "value_006"
  local_007 = "value_007"
  local_008 = "value_008"
  local_009 = "value_009"
  local_010 = "value_010"
  local_011 = "value_011"
  local_012 = "value_012"
  local_013 = "value_013"
  local_014 = "value_014"
  local_015 = "value_015"
  local_016 = "value_016"
  local_017 = "value_017"
  local_018 = "value_018"
  local_019 = "value_019"
  local_020 = "value_020"
}

resource "aws_s3_bucket" "bucket_001" {
  bucket = "my-bucket-001"
}

resource "aws_s3_bucket" "bucket_002" {
  bucket = "my-bucket-002"
}

resource "aws_s3_bucket" "bucket_003" {
  bucket = "my-bucket-003"
}

resource "aws_s3_bucket" "bucket_004" {
  bucket = "my-bucket-004"
}

resource "aws_s3_bucket" "bucket_005" {
  bucket = "my-bucket-005"
}

resource "aws_s3_bucket" "bucket_006" {
  bucket = "my-bucket-006"
}

resource "aws_s3_bucket" "bucket_007" {
  bucket = "my-bucket-007"
}

resource "aws_s3_bucket" "bucket_008" {
  bucket = "my-bucket-008"
}

resource "aws_s3_bucket" "bucket_009" {
  bucket = "my-bucket-009"
}

resource "aws_s3_bucket" "bucket_010" {
  bucket = "my-bucket-010"
}

resource "aws_s3_bucket" "bucket_011" {
  bucket = "my-bucket-011"
}

resource "aws_s3_bucket" "bucket_012" {
  bucket = "my-bucket-012"
}

resource "aws_s3_bucket" "bucket_013" {
  bucket = "my-bucket-013"
}

resource "aws_s3_bucket" "bucket_014" {
  bucket = "my-bucket-014"
}

resource "aws_s3_bucket" "bucket_015" {
  bucket = "my-bucket-015"
}

resource "aws_s3_bucket" "bucket_016" {
  bucket = "my-bucket-016"
}

resource "aws_s3_bucket" "bucket_017" {
  bucket = "my-bucket-017"
}

resource "aws_s3_bucket" "bucket_018" {
  bucket = "my-bucket-018"
}

resource "aws_s3_bucket" "bucket_019" {
  bucket = "my-bucket-019"
}

resource "aws_s3_bucket" "bucket_020" {
  bucket = "my-bucket-020"
}

resource "aws_s3_bucket" "bucket_021" {
  bucket = "my-bucket-021"
}

resource "aws_s3_bucket" "bucket_022" {
  bucket = "my-bucket-022"
}

resource "aws_s3_bucket" "bucket_023" {
  bucket = "my-bucket-023"
}

resource "aws_s3_bucket" "bucket_024" {
  bucket = "my-bucket-024"
}

resource "aws_s3_bucket" "bucket_025" {
  bucket = "my-bucket-025"
}

resource "aws_s3_bucket" "bucket_026" {
  bucket = "my-bucket-026"
}

resource "aws_s3_bucket" "bucket_027" {
  bucket = "my-bucket-027"
}

resource "aws_s3_bucket" "bucket_028" {
  bucket = "my-bucket-028"
}

resource "aws_s3_bucket" "bucket_029" {
  bucket = "my-bucket-029"
}

resource "aws_s3_bucket" "bucket_030" {
  bucket = "my-bucket-030"
}

resource "aws_s3_bucket" "bucket_031" {
  bucket = "my-bucket-031"
}

resource "aws_s3_bucket" "bucket_032" {
  bucket = "my-bucket-032"
}

resource "aws_s3_bucket" "bucket_033" {
  bucket = "my-bucket-033"
}

resource "aws_s3_bucket" "bucket_034" {
  bucket = "my-bucket-034"
}

resource "aws_s3_bucket" "bucket_035" {
  bucket = "my-bucket-035"
}

resource "aws_s3_bucket" "bucket_036" {
  bucket = "my-bucket-036"
}

resource "aws_s3_bucket" "bucket_037" {
  bucket = "my-bucket-037"
}

resource "aws_s3_bucket" "bucket_038" {
  bucket = "my-bucket-038"
}

resource "aws_s3_bucket" "bucket_039" {
  bucket = "my-bucket-039"
}

resource "aws_s3_bucket" "bucket_040" {
  bucket = "my-bucket-040"
}

resource "aws_s3_bucket" "bucket_041" {
  bucket = "my-bucket-041"
}

resource "aws_s3_bucket" "bucket_042" {
  bucket = "my-bucket-042"
}

resource "aws_s3_bucket" "bucket_043" {
  bucket = "my-bucket-043"
}

resource "aws_s3_bucket" "bucket_044" {
  bucket = "my-bucket-044"
}

resource "aws_s3_bucket" "bucket_045" {
  bucket = "my-bucket-045"
}

resource "aws_s3_bucket" "bucket_046" {
  bucket = "my-bucket-046"
}

resource "aws_s3_bucket" "bucket_047" {
  bucket = "my-bucket-047"
}

resource "aws_s3_bucket" "bucket_048" {
  bucket = "my-bucket-048"
}

resource "aws_s3_bucket" "bucket_049" {
  bucket = "my-bucket-049"
}

resource "aws_s3_bucket" "bucket_050" {
  bucket = "my-bucket-050"
}

resource "aws_s3_bucket" "bucket_051" {
  bucket = "my-bucket-051"
}

resource "aws_s3_bucket" "bucket_052" {
  bucket = "my-bucket-052"
}

resource "aws_s3_bucket" "bucket_053" {
  bucket = "my-bucket-053"
}

resource "aws_s3_bucket" "bucket_054" {
  bucket = "my-bucket-054"
}

resource "aws_s3_bucket" "bucket_055" {
  bucket = "my-bucket-055"
}

resource "aws_s3_bucket" "bucket_056" {
  bucket = "my-bucket-056"
}

resource "aws_s3_bucket" "bucket_057" {
  bucket = "my-bucket-057"
}

resource "aws_s3_bucket" "bucket_058" {
  bucket = "my-bucket-058"
}

resource "aws_s3_bucket" "bucket_059" {
  bucket = "my-bucket-059"
}

resource "aws_s3_bucket" "bucket_060" {
  bucket = "my-bucket-060"
}

resource "aws_s3_bucket" "bucket_061" {
  bucket = "my-bucket-061"
}

resource "aws_s3_bucket" "bucket_062" {
  bucket = "my-bucket-062"
}

resource "aws_s3_bucket" "bucket_063" {
  bucket = "my-bucket-063"
}

resource "aws_s3_bucket" "bucket_064" {
  bucket = "my-bucket-064"
}

resource "aws_s3_bucket" "bucket_065" {
  bucket = "my-bucket-065"
}

resource "aws_s3_bucket" "bucket_066" {
  bucket = "my-bucket-066"
}

resource "aws_s3_bucket" "bucket_067" {
  bucket = "my-bucket-067"
}

resource "aws_s3_bucket" "bucket_068" {
  bucket = "my-bucket-068"
}

resource "aws_s3_bucket" "bucket_069" {
  bucket = "my-bucket-069"
}

resource "aws_s3_bucket" "bucket_070" {
  bucket = "my-bucket-070"
}

resource "aws_s3_bucket" "bucket_071" {
  bucket = "my-bucket-071"
}

resource "aws_s3_bucket" "bucket_072" {
  bucket = "my-bucket-072"
}

resource "aws_s3_bucket" "bucket_073" {
  bucket = "my-bucket-073"
}

resource "aws_s3_bucket" "bucket_074" {
  bucket = "my-bucket-074"
}

resource "aws_s3_bucket" "bucket_075" {
  bucket = "my-bucket-075"
}

resource "aws_s3_bucket" "bucket_076" {
  bucket = "my-bucket-076"
}

resource "aws_s3_bucket" "bucket_077" {
  bucket = "my-bucket-077"
}

resource "aws_s3_bucket" "bucket_078" {
  bucket = "my-bucket-078"
}

resource "aws_s3_bucket" "bucket_079" {
  bucket = "my-bucket-079"
}

resource "aws_s3_bucket" "bucket_080" {
  bucket = "my-bucket-080"
}

resource "aws_s3_bucket" "bucket_081" {
  bucket = "my-bucket-081"
}

resource "aws_s3_bucket" "bucket_082" {
  bucket = "my-bucket-082"
}

resource "aws_s3_bucket" "bucket_083" {
  bucket = "my-bucket-083"
}

resource "aws_s3_bucket" "bucket_084" {
  bucket = "my-bucket-084"
}

resource "aws_s3_bucket" "bucket_085" {
  bucket = "my-bucket-085"
}

resource "aws_s3_bucket" "bucket_086" {
  bucket = "my-bucket-086"
}

resource "aws_s3_bucket" "bucket_087" {
  bucket = "my-bucket-087"
}

resource "aws_s3_bucket" "bucket_088" {
  bucket = "my-bucket-088"
}

resource "aws_s3_bucket" "bucket_089" {
  bucket = "my-bucket-089"
}

resource "aws_s3_bucket" "bucket_090" {
  bucket = "my-bucket-090"
}

resource "aws_s3_bucket" "bucket_091" {
  bucket = "my-bucket-091"
}

resource "aws_s3_bucket" "bucket_092" {
  bucket = "my-bucket-092"
}

resource "aws_s3_bucket" "bucket_093" {
  bucket = "my-bucket-093"
}

resource "aws_s3_bucket" "bucket_094" {
  bucket = "my-bucket-094"
}

resource "aws_s3_bucket" "bucket_095" {
  bucket = "my-bucket-095"
}

resource "aws_s3_bucket" "bucket_096" {
  bucket = "my-bucket-096"
}

resource "aws_s3_bucket" "bucket_097" {
  bucket = "my-bucket-097"
}

resource "aws_s3_bucket" "bucket_098" {
  bucket = "my-bucket-098"
}

resource "aws_s3_bucket" "bucket_099" {
  bucket = "my-bucket-099"
}

resource "aws_s3_bucket" "bucket_100" {
  bucket = "my-bucket-100"
}

resource "aws_s3_bucket" "bucket_101" {
  bucket = "my-bucket-101"
}

resource "aws_s3_bucket" "bucket_102" {
  bucket = "my-bucket-102"
}

resource "aws_s3_bucket" "bucket_103" {
  bucket = "my-bucket-103"
}

resource "aws_s3_bucket" "bucket_104" {
  bucket = "my-bucket-104"
}

resource "aws_s3_bucket" "bucket_105" {
  bucket = "my-bucket-105"
}

resource "aws_s3_bucket" "bucket_106" {
  bucket = "my-bucket-106"
}

resource "aws_s3_bucket" "bucket_107" {
  bucket = "my-bucket-107"
}

resource "aws_s3_bucket" "bucket_108" {
  bucket = "my-bucket-108"
}

resource "aws_s3_bucket" "bucket_109" {
  bucket = "my-bucket-109"
}

resource "aws_s3_bucket" "bucket_110" {
  bucket = "my-bucket-110"
}

resource "aws_s3_bucket" "bucket_111" {
  bucket = "my-bucket-111"
}

resource "aws_s3_bucket" "bucket_112" {
  bucket = "my-bucket-112"
}

resource "aws_s3_bucket" "bucket_113" {
  bucket = "my-bucket-113"
}

resource "aws_s3_bucket" "bucket_114" {
  bucket = "my-bucket-114"
}

resource "aws_s3_bucket" "bucket_115" {
  bucket = "my-bucket-115"
}

resource "aws_s3_bucket" "bucket_116" {
  bucket = "my-bucket-116"
}

resource "aws_s3_bucket" "bucket_117" {
  bucket = "my-bucket-117"
}

resource "aws_s3_bucket" "bucket_118" {
  bucket = "my-bucket-118"
}

resource "aws_s3_bucket" "bucket_119" {
  bucket = "my-bucket-119"
}

resource "aws_s3_bucket" "bucket_120" {
  bucket = "my-bucket-120"
}

resource "aws_s3_bucket" "bucket_121" {
  bucket = "my-bucket-121"
}

resource "aws_s3_bucket" "bucket_122" {
  bucket = "my-bucket-122"
}

resource "aws_s3_bucket" "bucket_123" {
  bucket = "my-bucket-123"
}

resource "aws_s3_bucket" "bucket_124" {
  bucket = "my-bucket-124"
}

resource "aws_s3_bucket" "bucket_125" {
  bucket = "my-bucket-125"
}

resource "aws_s3_bucket" "bucket_126" {
  bucket = "my-bucket-126"
}

resource "aws_s3_bucket" "bucket_127" {
  bucket = "my-bucket-127"
}

resource "aws_s3_bucket" "bucket_128" {
  bucket = "my-bucket-128"
}

resource "aws_s3_bucket" "bucket_129" {
  bucket = "my-bucket-129"
}

resource "aws_s3_bucket" "bucket_130" {
  bucket = "my-bucket-130"
}

resource "aws_s3_bucket" "bucket_131" {
  bucket = "my-bucket-131"
}

resource "aws_s3_bucket" "bucket_132" {
  bucket = "my-bucket-132"
}

resource "aws_s3_bucket" "bucket_133" {
  bucket = "my-bucket-133"
}

resource "aws_s3_bucket" "bucket_134" {
  bucket = "my-bucket-134"
}

resource "aws_s3_bucket" "bucket_135" {
  bucket = "my-bucket-135"
}

resource "aws_s3_bucket" "bucket_136" {
  bucket = "my-bucket-136"
}

resource "aws_s3_bucket" "bucket_137" {
  bucket = "my-bucket-137"
}

resource "aws_s3_bucket" "bucket_138" {
  bucket = "my-bucket-138"
}

resource "aws_s3_bucket" "bucket_139" {
  bucket = "my-bucket-139"
}

resource "aws_s3_bucket" "bucket_140" {
  bucket = "my-bucket-140"
}

resource "aws_s3_bucket" "bucket_141" {
  bucket = "my-bucket-141"
}

resource "aws_s3_bucket" "bucket_142" {
  bucket = "my-bucket-142"
}

resource "aws_s3_bucket" "bucket_143" {
  bucket = "my-bucket-143"
}

resource "aws_s3_bucket" "bucket_144" {
  bucket = "my-bucket-144"
}

resource "aws_s3_bucket" "bucket_145" {
  bucket = "my-bucket-145"
}

resource "aws_s3_bucket" "bucket_146" {
  bucket = "my-bucket-146"
}

resource "aws_s3_bucket" "bucket_147" {
  bucket = "my-bucket-147"
}

resource "aws_s3_bucket" "bucket_148" {
  bucket = "my-bucket-148"
}

resource "aws_s3_bucket" "bucket_149" {
  bucket = "my-bucket-149"
}

resource "aws_s3_bucket" "bucket_150" {
  bucket = "my-bucket-150"
}

resource "aws_s3_bucket" "bucket_151" {
  bucket = "my-bucket-151"
}

resource "aws_s3_bucket" "bucket_152" {
  bucket = "my-bucket-152"
}

resource "aws_s3_bucket" "bucket_153" {
  bucket = "my-bucket-153"
}

resource "aws_s3_bucket" "bucket_154" {
  bucket = "my-bucket-154"
}

resource "aws_s3_bucket" "bucket_155" {
  bucket = "my-bucket-155"
}

resource "aws_s3_bucket" "bucket_156" {
  bucket = "my-bucket-156"
}

resource "aws_s3_bucket" "bucket_157" {
  bucket = "my-bucket-157"
}

resource "aws_s3_bucket" "bucket_158" {
  bucket = "my-bucket-158"
}

resource "aws_s3_bucket" "bucket_159" {
  bucket = "my-bucket-159"
}

resource "aws_s3_bucket" "bucket_160" {
  bucket = "my-bucket-160"
}

resource "aws_s3_bucket" "bucket_161" {
  bucket = "my-bucket-161"
}

resource "aws_s3_bucket" "bucket_162" {
  bucket = "my-bucket-162"
}

resource "aws_s3_bucket" "bucket_163" {
  bucket = "my-bucket-163"
}

resource "aws_s3_bucket" "bucket_164" {
  bucket = "my-bucket-164"
}

resource "aws_s3_bucket" "bucket_165" {
  bucket = "my-bucket-165"
}

resource "aws_s3_bucket" "bucket_166" {
  bucket = "my-bucket-166"
}

resource "aws_s3_bucket" "bucket_167" {
  bucket = "my-bucket-167"
}

resource "aws_s3_bucket" "bucket_168" {
  bucket = "my-bucket-168"
}

resource "aws_s3_bucket" "bucket_169" {
  bucket = "my-bucket-169"
}

resource "aws_s3_bucket" "bucket_170" {
  bucket = "my-bucket-170"
}

resource "aws_s3_bucket" "bucket_171" {
  bucket = "my-bucket-171"
}

resource "aws_s3_bucket" "bucket_172" {
  bucket = "my-bucket-172"
}

resource "aws_s3_bucket" "bucket_173" {
  bucket = "my-bucket-173"
}

resource "aws_s3_bucket" "bucket_174" {
  bucket = "my-bucket-174"
}

resource "aws_s3_bucket" "bucket_175" {
  bucket = "my-bucket-175"
}

resource "aws_s3_bucket" "bucket_176" {
  bucket = "my-bucket-176"
}

resource "aws_s3_bucket" "bucket_177" {
  bucket = "my-bucket-177"
}

resource "aws_s3_bucket" "bucket_178" {
  bucket = "my-bucket-178"
}

resource "aws_s3_bucket" "bucket_179" {
  bucket = "my-bucket-179"
}

resource "aws_s3_bucket" "bucket_180" {
  bucket = "my-bucket-180"
}

resource "aws_s3_bucket" "bucket_181" {
  bucket = "my-bucket-181"
}

resource "aws_s3_bucket" "bucket_182" {
  bucket = "my-bucket-182"
}

resource "aws_s3_bucket" "bucket_183" {
  bucket = "my-bucket-183"
}

resource "aws_s3_bucket" "bucket_184" {
  bucket = "my-bucket-184"
}

resource "aws_s3_bucket" "bucket_185" {
  bucket = "my-bucket-185"
}

resource "aws_s3_bucket" "bucket_186" {
  bucket = "my-bucket-186"
}

resource "aws_s3_bucket" "bucket_187" {
  bucket = "my-bucket-187"
}

resource "aws_s3_bucket" "bucket_188" {
  bucket = "my-bucket-188"
}

resource "aws_s3_bucket" "bucket_189" {
  bucket = "my-bucket-189"
}

resource "aws_s3_bucket" "bucket_190" {
  bucket = "my-bucket-190"
}

resource "aws_s3_bucket" "bucket_191" {
  bucket = "my-bucket-191"
}

resource "aws_s3_bucket" "bucket_192" {
  bucket = "my-bucket-192"
}

resource "aws_s3_bucket" "bucket_193" {
  bucket = "my-bucket-193"
}

resource "aws_s3_bucket" "bucket_194" {
  bucket = "my-bucket-194"
}

resource "aws_s3_bucket" "bucket_195" {
  bucket = "my-bucket-195"
}

resource "aws_s3_bucket" "bucket_196" {
  bucket = "my-bucket-196"
}

resource "aws_s3_bucket" "bucket_197" {
  bucket = "my-bucket-197"
}

resource "aws_s3_bucket" "bucket_198" {
  bucket = "my-bucket-198"
}

resource "aws_s3_bucket" "bucket_199" {
  bucket = "my-bucket-199"
}

resource "aws_s3_bucket" "bucket_200" {
  bucket = "my-bucket-200"
}

resource "aws_instance" "instance_001" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-001"
  }
}

resource "aws_instance" "instance_002" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-002"
  }
}

resource "aws_instance" "instance_003" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-003"
  }
}

resource "aws_instance" "instance_004" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-004"
  }
}

resource "aws_instance" "instance_005" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-005"
  }
}

resource "aws_instance" "instance_006" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-006"
  }
}

resource "aws_instance" "instance_007" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-007"
  }
}

resource "aws_instance" "instance_008" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-008"
  }
}

resource "aws_instance" "instance_009" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-009"
  }
}

resource "aws_instance" "instance_010" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-010"
  }
}

resource "aws_instance" "instance_011" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-011"
  }
}

resource "aws_instance" "instance_012" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-012"
  }
}

resource "aws_instance" "instance_013" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-013"
  }
}

resource "aws_instance" "instance_014" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-014"
  }
}

resource "aws_instance" "instance_015" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-015"
  }
}

resource "aws_instance" "instance_016" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-016"
  }
}

resource "aws_instance" "instance_017" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-017"
  }
}

resource "aws_instance" "instance_018" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-018"
  }
}

resource "aws_instance" "instance_019" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-019"
  }
}

resource "aws_instance" "instance_020" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-020"
  }
}

resource "aws_instance" "instance_021" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-021"
  }
}

resource "aws_instance" "instance_022" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-022"
  }
}

resource "aws_instance" "instance_023" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-023"
  }
}

resource "aws_instance" "instance_024" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-024"
  }
}

resource "aws_instance" "instance_025" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-025"
  }
}

resource "aws_instance" "instance_026" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-026"
  }
}

resource "aws_instance" "instance_027" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-027"
  }
}

resource "aws_instance" "instance_028" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-028"
  }
}

resource "aws_instance" "instance_029" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-029"
  }
}

resource "aws_instance" "instance_030" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-030"
  }
}

resource "aws_instance" "instance_031" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-031"
  }
}

resource "aws_instance" "instance_032" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-032"
  }
}

resource "aws_instance" "instance_033" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-033"
  }
}

resource "aws_instance" "instance_034" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-034"
  }
}

resource "aws_instance" "instance_035" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-035"
  }
}

resource "aws_instance" "instance_036" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-036"
  }
}

resource "aws_instance" "instance_037" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-037"
  }
}

resource "aws_instance" "instance_038" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-038"
  }
}

resource "aws_instance" "instance_039" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-039"
  }
}

resource "aws_instance" "instance_040" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-040"
  }
}

resource "aws_instance" "instance_041" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-041"
  }
}

resource "aws_instance" "instance_042" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-042"
  }
}

resource "aws_instance" "instance_043" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-043"
  }
}

resource "aws_instance" "instance_044" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-044"
  }
}

resource "aws_instance" "instance_045" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-045"
  }
}

resource "aws_instance" "instance_046" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-046"
  }
}

resource "aws_instance" "instance_047" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-047"
  }
}

resource "aws_instance" "instance_048" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-048"
  }
}

resource "aws_instance" "instance_049" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-049"
  }
}

resource "aws_instance" "instance_050" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-050"
  }
}

resource "aws_instance" "instance_051" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-051"
  }
}

resource "aws_instance" "instance_052" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-052"
  }
}

resource "aws_instance" "instance_053" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-053"
  }
}

resource "aws_instance" "instance_054" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-054"
  }
}

resource "aws_instance" "instance_055" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-055"
  }
}

resource "aws_instance" "instance_056" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-056"
  }
}

resource "aws_instance" "instance_057" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-057"
  }
}

resource "aws_instance" "instance_058" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-058"
  }
}

resource "aws_instance" "instance_059" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-059"
  }
}

resource "aws_instance" "instance_060" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-060"
  }
}

resource "aws_instance" "instance_061" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-061"
  }
}

resource "aws_instance" "instance_062" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-062"
  }
}

resource "aws_instance" "instance_063" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-063"
  }
}

resource "aws_instance" "instance_064" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-064"
  }
}

resource "aws_instance" "instance_065" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-065"
  }
}

resource "aws_instance" "instance_066" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-066"
  }
}

resource "aws_instance" "instance_067" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-067"
  }
}

resource "aws_instance" "instance_068" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-068"
  }
}

resource "aws_instance" "instance_069" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-069"
  }
}

resource "aws_instance" "instance_070" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-070"
  }
}

resource "aws_instance" "instance_071" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-071"
  }
}

resource "aws_instance" "instance_072" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-072"
  }
}

resource "aws_instance" "instance_073" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-073"
  }
}

resource "aws_instance" "instance_074" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-074"
  }
}

resource "aws_instance" "instance_075" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-075"
  }
}

resource "aws_instance" "instance_076" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-076"
  }
}

resource "aws_instance" "instance_077" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-077"
  }
}

resource "aws_instance" "instance_078" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-078"
  }
}

resource "aws_instance" "instance_079" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-079"
  }
}

resource "aws_instance" "instance_080" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-080"
  }
}

resource "aws_instance" "instance_081" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-081"
  }
}

resource "aws_instance" "instance_082" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-082"
  }
}

resource "aws_instance" "instance_083" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-083"
  }
}

resource "aws_instance" "instance_084" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-084"
  }
}

resource "aws_instance" "instance_085" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-085"
  }
}

resource "aws_instance" "instance_086" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-086"
  }
}

resource "aws_instance" "instance_087" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-087"
  }
}

resource "aws_instance" "instance_088" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-088"
  }
}

resource "aws_instance" "instance_089" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-089"
  }
}

resource "aws_instance" "instance_090" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-090"
  }
}

resource "aws_instance" "instance_091" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-091"
  }
}

resource "aws_instance" "instance_092" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-092"
  }
}

resource "aws_instance" "instance_093" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-093"
  }
}

resource "aws_instance" "instance_094" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-094"
  }
}

resource "aws_instance" "instance_095" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-095"
  }
}

resource "aws_instance" "instance_096" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-096"
  }
}

resource "aws_instance" "instance_097" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-097"
  }
}

resource "aws_instance" "instance_098" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-098"
  }
}

resource "aws_instance" "instance_099" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-099"
  }
}

resource "aws_instance" "instance_100" {
  ami = "ami-12345678"
  instance_type = "t3.micro"
  tags = {
    Name = "instance-100"
  }
}

output "output_001" {
  value = aws_instance.instance_001.id
}

output "output_002" {
  value = aws_instance.instance_002.id
}

output "output_003" {
  value = aws_instance.instance_003.id
}

output "output_004" {
  value = aws_instance.instance_004.id
}

output "output_005" {
  value = aws_instance.instance_005.id
}

output "output_006" {
  value = aws_instance.instance_006.id
}

output "output_007" {
  value = aws_instance.instance_007.id
}

output "output_008" {
  value = aws_instance.instance_008.id
}

output "output_009" {
  value = aws_instance.instance_009.id
}

output "output_010" {
  value = aws_instance.instance_010.id
}

output "output_011" {
  value = aws_instance.instance_011.id
}

output "output_012" {
  value = aws_instance.instance_012.id
}

output "output_013" {
  value = aws_instance.instance_013.id
}

output "output_014" {
  value = aws_instance.instance_014.id
}

output "output_015" {
  value = aws_instance.instance_015.id
}

output "output_016" {
  value = aws_instance.instance_016.id
}

output "output_017" {
  value = aws_instance.instance_017.id
}

output "output_018" {
  value = aws_instance.instance_018.id
}

output "output_019" {
  value = aws_instance.instance_019.id
}

output "output_020" {
  value = aws_instance.instance_020.id
}

output "output_021" {
  value = aws_instance.instance_021.id
}

output "output_022" {
  value = aws_instance.instance_022.id
}

output "output_023" {
  value = aws_instance.instance_023.id
}

output "output_024" {
  value = aws_instance.instance_024.id
}

output "output_025" {
  value = aws_instance.instance_025.id
}

output "output_026" {
  value = aws_instance.instance_026.id
}

output "output_027" {
  value = aws_instance.instance_027.id
}

output "output_028" {
  value = aws_instance.instance_028.id
}

output "output_029" {
  value = aws_instance.instance_029.id
}

output "output_030" {
  value = aws_instance.instance_030.id
}

output "output_031" {
  value = aws_instance.instance_031.id
}

output "output_032" {
  value = aws_instance.instance_032.id
}

output "output_033" {
  value = aws_instance.instance_033.id
}

output "output_034" {
  value = aws_instance.instance_034.id
}

output "output_035" {
  value = aws_instance.instance_035.id
}

output "output_036" {
  value = aws_instance.instance_036.id
}

output "output_037" {
  value = aws_instance.instance_037.id
}

output "output_038" {
  value = aws_instance.instance_038.id
}

output "output_039" {
  value = aws_instance.instance_039.id
}

output "output_040" {
  value = aws_instance.instance_040.id
}

output "output_041" {
  value = aws_instance.instance_041.id
}

output "output_042" {
  value = aws_instance.instance_042.id
}

output "output_043" {
  value = aws_instance.instance_043.id
}

output "output_044" {
  value = aws_instance.instance_044.id
}

output "output_045" {
  value = aws_instance.instance_045.id
}

output "output_046" {
  value = aws_instance.instance_046.id
}

output "output_047" {
  value = aws_instance.instance_047.id
}

output "output_048" {
  value = aws_instance.instance_048.id
}

output "output_049" {
  value = aws_instance.instance_049.id
}

output "output_050" {
  value = aws_instance.instance_050.id
}

output "output_051" {
  value = aws_instance.instance_051.id
}

output "output_052" {
  value = aws_instance.instance_052.id
}

output "output_053" {
  value = aws_instance.instance_053.id
}

output "output_054" {
  value = aws_instance.instance_054.id
}

output "output_055" {
  value = aws_instance.instance_055.id
}

output "output_056" {
  value = aws_instance.instance_056.id
}

output "output_057" {
  value = aws_instance.instance_057.id
}

output "output_058" {
  value = aws_instance.instance_058.id
}

output "output_059" {
  value = aws_instance.instance_059.id
}

output "output_060" {
  value = aws_instance.instance_060.id
}

output "output_061" {
  value = aws_instance.instance_061.id
}

output "output_062" {
  value = aws_instance.instance_062.id
}

output "output_063" {
  value = aws_instance.instance_063.id
}

output "output_064" {
  value = aws_instance.instance_064.id
}

output "output_065" {
  value = aws_instance.instance_065.id
}

output "output_066" {
  value = aws_instance.instance_066.id
}

output "output_067" {
  value = aws_instance.instance_067.id
}

output "output_068" {
  value = aws_instance.instance_068.id
}

output "output_069" {
  value = aws_instance.instance_069.id
}

output "output_070" {
  value = aws_instance.instance_070.id
}

output "output_071" {
  value = aws_instance.instance_071.id
}

output "output_072" {
  value = aws_instance.instance_072.id
}

output "output_073" {
  value = aws_instance.instance_073.id
}

output "output_074" {
  value = aws_instance.instance_074.id
}

output "output_075" {
  value = aws_instance.instance_075.id
}

output "output_076" {
  value = aws_instance.instance_076.id
}

output "output_077" {
  value = aws_instance.instance_077.id
}

output "output_078" {
  value = aws_instance.instance_078.id
}

output "output_079" {
  value = aws_instance.instance_079.id
}

output "output_080" {
  value = aws_instance.instance_080.id
}

output "output_081" {
  value = aws_instance.instance_081.id
}

output "output_082" {
  value = aws_instance.instance_082.id
}

output "output_083" {
  value = aws_instance.instance_083.id
}

output "output_084" {
  value = aws_instance.instance_084.id
}

output "output_085" {
  value = aws_instance.instance_085.id
}

output "output_086" {
  value = aws_instance.instance_086.id
}

output "output_087" {
  value = aws_instance.instance_087.id
}

output "output_088" {
  value = aws_instance.instance_088.id
}

output "output_089" {
  value = aws_instance.instance_089.id
}

output "output_090" {
  value = aws_instance.instance_090.id
}

output "output_091" {
  value = aws_instance.instance_091.id
}

output "output_092" {
  value = aws_instance.instance_092.id
}

output "output_093" {
  value = aws_instance.instance_093.id
}

output "output_094" {
  value = aws_instance.instance_094.id
}

output "output_095" {
  value = aws_instance.instance_095.id
}

output "output_096" {
  value = aws_instance.instance_096.id
}

output "output_097" {
  value = aws_instance.instance_097.id
}

output "output_098" {
  value = aws_instance.instance_098.id
}

output "output_099" {
  value = aws_instance.instance_099.id
}

output "output_100" {
  value = aws_instance.instance_100.id
}
