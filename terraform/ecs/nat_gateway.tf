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
