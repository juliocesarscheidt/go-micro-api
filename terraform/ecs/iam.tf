resource "aws_iam_role" "ecs_task_role" {
  name               = "AmazonECSTaskRole"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {
      "Service": ["ecs-tasks.amazonaws.com"]
    },
    "Action": "sts:AssumeRole",
    "Sid": ""
  }]
}
EOF
}

# required policy in order to allow enable_execute_command in the ecs service
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
  "Statement": [{
    "Action": "sts:AssumeRole",
    "Principal": {
      "Service": ["ecs-tasks.amazonaws.com"]
    },
    "Effect": "Allow",
    "Sid": ""
  }]
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
