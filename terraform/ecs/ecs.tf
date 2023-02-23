resource "aws_ecs_cluster" "ecs_cluster" {
  name               = "${var.api_name}-cluster"
  capacity_providers = ["FARGATE_SPOT", "FARGATE"]
  default_capacity_provider_strategy {
    capacity_provider = "FARGATE_SPOT"
  }
  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_cloudwatch_log_group" "container_log_group" {
  retention_in_days = 1
  name              = "/aws/ecs/${var.api_name}"
}

resource "aws_ecs_task_definition" "container_task_def" {
  family = var.api_name
  # role for task execution, which will be used to pull the image, create log stream, start the task, etc
  execution_role_arn = aws_iam_role.ecs_task_execution_role.arn
  # role for task application, to be used by the application itself in execution time, it's optional
  task_role_arn = aws_iam_role.ecs_task_role.arn
  container_definitions = jsonencode([
    {
      name : var.api_name
      image : "${var.registry_url}/${var.api_name}:${var.api_version}",
      portMappings = [
        {
          containerPort = var.api_port
          hostPort      = var.api_port
        }
      ],
      environment : [
        { "name" : "MESSAGE", "value" : var.api_message },
      ],
      cpu : tonumber(var.api_cpu),
      memory : tonumber(var.api_memory),
      essential : true,
      logConfiguration = {
        logDriver = "awslogs",
        Options = {
          "awslogs-region"        = var.aws_region,
          "awslogs-group"         = aws_cloudwatch_log_group.container_log_group.name,
          "awslogs-stream-prefix" = "ecs",
        }
      },
    },
  ])
  network_mode             = "awsvpc"
  cpu                      = tonumber(var.api_cpu)
  memory                   = tonumber(var.api_memory)
  requires_compatibilities = ["FARGATE"]
  depends_on = [
    aws_cloudwatch_log_group.container_log_group,
  ]
}

resource "aws_ecs_service" "container_service" {
  name                               = var.api_name
  cluster                            = aws_ecs_cluster.ecs_cluster.name
  task_definition                    = aws_ecs_task_definition.container_task_def.arn
  scheduling_strategy                = "REPLICA"
  launch_type                        = "FARGATE"
  desired_count                      = 1
  deployment_minimum_healthy_percent = "100"
  deployment_maximum_percent         = "200"
  enable_execute_command             = true
  force_new_deployment               = false
  network_configuration {
    subnets          = aws_subnet.private_subnet[*].id
    security_groups  = [aws_security_group.api_sg.id]
    assign_public_ip = false
  }
  load_balancer {
    target_group_arn = aws_alb_target_group.container_tg.arn
    container_name   = var.api_name
    container_port   = var.api_port
  }
  depends_on = [
    aws_ecs_cluster.ecs_cluster,
    aws_alb_target_group.container_tg,
    aws_security_group.api_sg,
  ]
}

resource "aws_iam_role" "ecs_task_role" {
  name               = "AmazonECSTaskRole"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {"Service": ["ecs-tasks.amazonaws.com"]},
      "Action": "sts:AssumeRole",
      "Sid": ""
    }
  ]
}
EOF
}

# required policies in order to allow enable_execute_command in the service
resource "aws_iam_policy" "ecs_task_role_policy" {
  name   = "AmazonECSTaskRolePolicy"
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": [
      "ssmmessages:CreateControlChannel",
      "ssmmessages:CreateDataChannel",
      "ssmmessages:OpenControlChannel",
      "ssmmessages:OpenDataChannel"
    ],
    "Resource": "*"
  }]
}
EOF
}

resource "aws_iam_role_policy_attachment" "attach_ecs_task_role_policy" {
  role       = aws_iam_role.ecs_task_role.name
  policy_arn = aws_iam_policy.ecs_task_role_policy.arn
  depends_on = [
    aws_iam_role.ecs_task_role,
    aws_iam_policy.ecs_task_role_policy,
  ]
}

resource "aws_iam_role" "ecs_task_execution_role" {
  name               = "AmazonECSTaskExecutionRole"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {"Service": ["ecs-tasks.amazonaws.com"]},
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "attach_AmazonECSTaskExecutionRolePolicy" {
  role       = aws_iam_role.ecs_task_execution_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

resource "aws_iam_role_policy_attachment" "attach_AmazonECS_FullAccess" {
  role       = aws_iam_role.ecs_task_execution_role.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonECS_FullAccess"
}
