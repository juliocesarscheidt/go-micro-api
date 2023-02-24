resource "aws_security_group" "alb_sg" {
  vpc_id = aws_vpc.vpc_0.id
  name   = "${var.api_name}-alb-sg"
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
  lifecycle {
    create_before_destroy = true
  }
  depends_on = [aws_vpc.vpc_0]
}

resource "aws_security_group" "api_sg" {
  vpc_id = aws_vpc.vpc_0.id
  name   = "${var.api_name}-sg"
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
  ingress {
    from_port       = var.api_port
    to_port         = var.api_port
    protocol        = "tcp"
    security_groups = [aws_security_group.alb_sg.id]
  }
  lifecycle {
    create_before_destroy = true
  }
  depends_on = [aws_security_group.alb_sg]
}
