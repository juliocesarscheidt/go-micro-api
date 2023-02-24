data "aws_availability_zones" "available_azs" {
  state = "available"
}

locals {
  subnets_offset  = min(length(data.aws_availability_zones.available_azs.names), 2)
  azs_index       = ["a", "b"]
  public_subnets  = [for index in range(0, local.subnets_offset) : cidrsubnet(aws_vpc.vpc_0.cidr_block, 4, index)]
  private_subnets = [for index in range(0, local.subnets_offset) : cidrsubnet(aws_vpc.vpc_0.cidr_block, 4, index + local.subnets_offset)]
}

resource "aws_vpc" "vpc_0" {
  cidr_block           = "10.100.0.0/20"
  instance_tenancy     = "default"
  enable_dns_support   = "true"
  enable_dns_hostnames = "true"
  enable_classiclink   = "false"
  tags = {
    Name = "${var.api_name}-vpc"
  }
}

######## public subnets ########
resource "aws_subnet" "public_subnet" {
  count                   = local.subnets_offset
  cidr_block              = local.public_subnets[count.index]
  availability_zone       = data.aws_availability_zones.available_azs.names[count.index]
  vpc_id                  = aws_vpc.vpc_0.id
  map_public_ip_on_launch = true
  tags = {
    Name = "${var.api_name}-public-${local.azs_index[count.index]}"
  }
  depends_on = [aws_vpc.vpc_0]
}

resource "aws_route_table" "public_route_table" {
  vpc_id = aws_vpc.vpc_0.id
  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.internet_gw.id
  }
  tags = {
    Name = "${var.api_name}-public-rt"
  }
  depends_on = [aws_internet_gateway.internet_gw]
}

resource "aws_route_table_association" "assoc_route_public" {
  count          = local.subnets_offset
  subnet_id      = element(aws_subnet.public_subnet[*].id, count.index)
  route_table_id = aws_route_table.public_route_table.id
  depends_on     = [aws_subnet.public_subnet, aws_route_table.public_route_table]
}

# change the main route
resource "aws_main_route_table_association" "assoc_main_route" {
  vpc_id         = aws_vpc.vpc_0.id
  route_table_id = aws_route_table.public_route_table.id
  depends_on     = [aws_vpc.vpc_0, aws_route_table.public_route_table]
}

######## private subnets ########
resource "aws_subnet" "private_subnet" {
  count                   = local.subnets_offset
  cidr_block              = local.private_subnets[count.index]
  availability_zone       = data.aws_availability_zones.available_azs.names[count.index]
  vpc_id                  = aws_vpc.vpc_0.id
  map_public_ip_on_launch = false
  tags = {
    Name = "${var.api_name}-private-${local.azs_index[count.index]}"
  }
  depends_on = [aws_vpc.vpc_0]
}

resource "aws_route_table" "private_route_table" {
  vpc_id = aws_vpc.vpc_0.id
  route {
    cidr_block     = "0.0.0.0/0"
    nat_gateway_id = aws_nat_gateway.nat_gw.id
  }
  tags = {
    Name = "${var.api_name}-private-rt"
  }
  depends_on = [aws_nat_gateway.nat_gw]
}

resource "aws_route_table_association" "assoc_route_private" {
  count          = local.subnets_offset
  subnet_id      = element(aws_subnet.private_subnet[*].id, count.index)
  route_table_id = aws_route_table.private_route_table.id
  depends_on     = [aws_subnet.private_subnet, aws_route_table.private_route_table]
}

######## internet and nat gateways ########
resource "aws_internet_gateway" "internet_gw" {
  vpc_id = aws_vpc.vpc_0.id
  tags = {
    Name = "${var.api_name}-int-gw"
  }
  depends_on = [aws_vpc.vpc_0]
}

resource "aws_eip" "nat_eip" {
  vpc = true
}

resource "aws_nat_gateway" "nat_gw" {
  allocation_id = aws_eip.nat_eip.id
  subnet_id     = aws_subnet.public_subnet[0].id
  tags = {
    Name = "${var.api_name}-nat-gw"
  }
  depends_on = [aws_eip.nat_eip, aws_subnet.public_subnet]
}

