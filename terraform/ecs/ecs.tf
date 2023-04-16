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

resource "aws_cloudwatch_log_group" "api_container_log_group" {
  retention_in_days = 1
  name              = "/aws/ecs/${var.api_name}"
}

resource "aws_ecs_task_definition" "api_container_task_def" {
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
        { "name" : "ENVIRONMENT", "value" : var.api_environment },
      ],
      cpu : tonumber(var.api_cpu),
      memory : tonumber(var.api_memory),
      essential : true,
      logConfiguration = {
        logDriver = "awslogs",
        Options = {
          "awslogs-region"        = var.aws_region,
          "awslogs-group"         = aws_cloudwatch_log_group.api_container_log_group.name,
          "awslogs-stream-prefix" = "ecs",
        }
      },
    },
  ])
  network_mode             = "awsvpc"
  cpu                      = tonumber(var.api_cpu)
  memory                   = tonumber(var.api_memory)
  requires_compatibilities = ["FARGATE"]
  tags = {
    "Name" = var.api_name
  }
  depends_on = [
    aws_cloudwatch_log_group.api_container_log_group,
  ]
}

resource "aws_ecs_service" "api_container_service" {
  name                               = var.api_name
  cluster                            = aws_ecs_cluster.ecs_cluster.name
  task_definition                    = aws_ecs_task_definition.api_container_task_def.arn
  scheduling_strategy                = "REPLICA"
  launch_type                        = "FARGATE"
  desired_count                      = var.api_replicas_count
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
  tags = {
    "Name"    = var.api_name
    "Cluster" = aws_ecs_cluster.ecs_cluster.name
  }
  depends_on = [
    aws_ecs_cluster.ecs_cluster,
    aws_alb_target_group.container_tg,
    aws_security_group.api_sg,
  ]
}
