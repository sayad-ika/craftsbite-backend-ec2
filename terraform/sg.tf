resource "aws_security_group" "craftsbite_sg"	{
    name = "${var.app_name}-sg"
    description = "Secuirty Group for Craftsbite server"

    ingress {
	description = "SSH"
	from_port = 22
	to_port = 22
	protocol = "tcp"
	cidr_blocks = ["0.0.0.0/0"]
    }

    ingress {
	description = "HTTP"
	from_port = 80
	to_port = 80
	protocol = "tcp"
	cidr_blocks = ["0.0.0.0/0"]
    }

    egress  {
	from_port = 0
	to_port = 0
	protocol = "-1"
	cidr_blocks = ["0.0.0.0/0"]
    }

    tags = {
	Name = "${var.app_name}-sg"
    }
}

