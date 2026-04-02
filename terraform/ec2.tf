resource "aws_instance" "craftsbite" {
  ami                    = var.ami_id
  instance_type          = var.instance_type
  key_name               = var.key_name
  vpc_security_group_ids = [aws_security_group.craftsbite_sg.id]
  iam_instance_profile   = data.aws_iam_instance_profile.ec2_profile.name

  root_block_device {
    volume_size = 8
    volume_type = "gp3"
  }

  user_data = file("${path.module}/userdata.sh")

  tags = {
    Name = "${var.app_name}-server"
  }
}

# resource "aws_instance" "craftsbite"	{
#     ami = var.ami_id
#     instance_type = var.instance_type
#     key_name = var.key_name
#     vpc_security_group_ids = [aws_security_group.craftsbite_sg.id]
#     iam_instance_profile = aws_iam_instance_profile.ec2_profile.name
    
#     user_data = file("${path.module}/userdata.sh")

#     tags = {
# 	Name = "${var.app_name}-server"
#     }
# }

