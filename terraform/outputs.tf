output "public_ip" {
  value = aws_instance.craftsbite.public_ip
}

output "public_url" {
  value = "http://${aws_instance.craftsbite.public_ip}"
}

output "swagger_url" {
  value = "http://${aws_instance.craftsbite.public_ip}/swagger/index.html"
}

output "ssh_command" {
  value = "ssh -i ${var.key_name}.pem ec2-user@${aws_instance.craftsbite.public_ip}"
}