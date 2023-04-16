resource "aws_lb" "application_lb" {
  load_balancer_type         = "application"
  name                       = "${var.api_name}-alb"
  internal                   = false
  enable_deletion_protection = false
  idle_timeout               = 60
  subnets                    = aws_subnet.public_subnet[*].id
  security_groups            = [aws_security_group.alb_sg.id]
  tags = {
    Name = "${var.api_name}-alb"
  }
  lifecycle {
    create_before_destroy = true
  }
  depends_on = [
    aws_security_group.alb_sg,
    aws_subnet.public_subnet,
  ]
}

output "public_ip" {
  value = aws_lb.application_lb.dns_name
}

resource "aws_alb_target_group" "container_tg" {
  name                          = "${var.api_name}-tg"
  port                          = var.api_port
  protocol                      = "HTTP"
  vpc_id                        = aws_vpc.vpc_0.id
  load_balancing_algorithm_type = "round_robin" # round_robin or least_outstanding_requests
  deregistration_delay          = 30
  target_type                   = "ip"
  health_check {
    healthy_threshold   = 2
    unhealthy_threshold = 5
    timeout             = 10
    interval            = 15
    protocol            = "HTTP"
    path                = var.api_liveness_path
    port                = var.api_port
  }
  lifecycle {
    create_before_destroy = true
  }
  depends_on = [
    aws_vpc.vpc_0,
  ]
}

resource "aws_alb_listener" "application_lb_listener_http" {
  load_balancer_arn = aws_lb.application_lb.arn
  port              = 80
  protocol          = "HTTP"
  default_action {
    type             = "forward"
    target_group_arn = aws_alb_target_group.container_tg.id
  }
}
