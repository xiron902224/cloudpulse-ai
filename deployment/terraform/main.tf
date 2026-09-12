# ============================================
# Infrastructure as Code (IaC)
# CloudPulse AI
# ============================================
# 
# NOTE: This file is for when you use AWS/GCP/Azure
# You do not need it with the free tier for now,
# but documenting it is good practice
#
# ============================================

terraform {
  required_version = ">= 1.0"
  
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
  
  # Remote state (optional)
  # backend "s3" {
  #   bucket = "cloudpulse-terraform-state"
  #   key    = "prod/terraform.tfstate"
  #   region = "us-east-1"
  # }
}

# AWS provider
provider "aws" {
  region = var.aws_region
}

# Variables
variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "environment" {
  description = "Environment (dev, staging, prod)"
  type        = string
  default     = "dev"
}

variable "app_name" {
  description = "Application name"
  type        = string
  default     = "cloudpulse-ai"
}

# ============================================
# NETWORK (VPC)
# ============================================

resource "aws_vpc" "main" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name        = "${var.app_name}-vpc"
    Environment = var.environment
  }
}

resource "aws_subnet" "public_1" {
  vpc_id                  = aws_vpc.main.id
  cidr_block              = "10.0.1.0/24"
  availability_zone       = "${var.aws_region}a"
  map_public_ip_on_launch = true

  tags = {
    Name = "${var.app_name}-subnet-public-1"
  }
}

resource "aws_subnet" "public_2" {
  vpc_id                  = aws_vpc.main.id
  cidr_block              = "10.0.2.0/24"
  availability_zone       = "${var.aws_region}b"
  map_public_ip_on_launch = true

  tags = {
    Name = "${var.app_name}-subnet-public-2"
  }
}

resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id

  tags = {
    Name = "${var.app_name}-igw"
  }
}

resource "aws_route_table" "main" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.main.id
  }

  tags = {
    Name = "${var.app_name}-route-table"
  }
}

# ============================================
# SECURITY (Security Groups)
# ============================================

resource "aws_security_group" "api" {
  name        = "${var.app_name}-api-sg"
  description = "Security group for the API"
  vpc_id      = aws_vpc.main.id

  ingress {
    from_port   = 8080
    to_port     = 8080
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "HTTP API"
  }

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "SSH"
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "${var.app_name}-api-sg"
  }
}

# ============================================
# DATABASE (RDS PostgreSQL)
# ============================================

resource "aws_db_instance" "postgres" {
  identifier     = "${var.app_name}-postgres"
  engine         = "postgres"
  engine_version = "16.1"
  instance_class = "db.t3.micro"  # Gratis tier

  allocated_storage     = 20
  max_allocated_storage = 100

  db_name  = "cloudpulse"
  username = "admin"
  password = var.db_password

  publicly_accessible    = true
  skip_final_snapshot    = true
  vpc_security_group_ids = [aws_security_group.api.id]

  tags = {
    Name        = "${var.app_name}-postgres"
    Environment = var.environment
  }
}

variable "db_password" {
  description = "Database password"
  type        = string
  sensitive   = true
}

# ============================================
# CACHE (ElastiCache Redis)
# ============================================

resource "aws_elasticache_cluster" "redis" {
  cluster_id           = "${var.app_name}-redis"
  engine               = "redis"
  node_type            = "cache.t3.micro"  # Gratis tier
  num_cache_nodes      = 1
  parameter_group_name = "default.redis7"
  port                 = 6379

  tags = {
    Name        = "${var.app_name}-redis"
    Environment = var.environment
  }
}

# ============================================
# COMPUTE (EC2 for API)
# ============================================

resource "aws_instance" "api" {
  ami           = "ami-0c7217cdde317cfec"  # Ubuntu 22.04 LTS
  instance_type = "t2.micro"               # Gratis tier

  vpc_security_group_ids = [aws_security_group.api.id]
  subnet_id              = aws_subnet.public_1.id

  user_data = <<-EOF
    #!/bin/bash
    # Startup script
    echo "Starting API..."
    # API deployment would go here
  EOF

  tags = {
    Name        = "${var.app_name}-api"
    Environment = var.environment
  }
}

# ============================================
# OUTPUTS
# ============================================

output "api_url" {
  description = "API URL"
  value       = "http://${aws_instance.api.public_ip}:8080"
}

output "database_endpoint" {
  description = "PostgreSQL endpoint"
  value       = aws_db_instance.postgres.endpoint
}

output "redis_endpoint" {
  description = "Redis endpoint"
  value       = aws_elasticache_cluster.redis.cache_nodes[0].address
}