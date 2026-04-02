#!/bin/bash
set -e

S3_BUCKET="trainee-2026-sayad-craftsbite"
BINARY_NAME="craftsbite-server"
APP_DIR="/home/ec2-user/craftsbite-backend-ec2"
SERVICE_NAME="craftsbite"
APP_PORT="8080"

# System Update
echo "🔄 Updating System..."
dnf update -y

echo "Installing Nginx"
dnf install nginx -y
systemctl enable nginx
systemctl start nginx

echo "Creating app dir"
mkdir -p $APP_DIR
chown ec2-user:ec2-user $APP_DIR

echo "Downloading binary & .env from s3 bucket"
aws s3 cp s3://$S3_BUCKET/$BINARY_NAME $APP_DIR
chmod +x $APP_DIR/$BINARY_NAME
chown ec2-user:ec2-user $APP_DIR/$BINARY_NAME

aws s3 cp s3://$S3_BUCKET/.env $APP_DIR
chown ec2-user:ec2-user $APP_DIR/.env

echo "Creating systemd service..."
cat > /etc/systemd/system/$SERVICE_NAME.service <<EOF
[Unit]
Description=Craftsbite API Server
After=network.target

[Service]
User=ec2-user
WorkingDirectory=$APP_DIR
ExecStart=$APP_DIR/$BINARY_NAME
Restart=always
RestartSec=5
EnvironmentFile=$APP_DIR/.env

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable $SERVICE_NAME
systemctl start $SERVICE_NAME

echo "Configuring NGINX"
# Fix: Replace nginx.conf to remove default server block conflict
cat > /etc/nginx/nginx.conf <<'EOF'
user nginx;
worker_processes auto;
error_log /var/log/nginx/error.log notice;
pid /run/nginx.pid;

include /usr/share/nginx/modules/*.conf;

events {
    worker_connections 1024;
}

http {
    log_format  main  '$remote_addr - $remote_user [$time_local] "$request" '
                      '$status $body_bytes_sent "$http_referer" '
                      '"$http_user_agent" "$http_x_forwarded_for"';

    access_log  /var/log/nginx/access.log  main;

    sendfile            on;
    tcp_nopush          on;
    keepalive_timeout   65;
    types_hash_max_size 4096;

    include             /etc/nginx/mime.types;
    default_type        application/octet-stream;

    include /etc/nginx/conf.d/*.conf;
}
EOF

cat > /etc/nginx/conf.d/craftsbite.conf <<EOF
server {
    listen 80;

    location / {
        proxy_pass http://localhost:$APP_PORT;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
    }
}
EOF

nginx -t
systemctl restart nginx

echo "Creating deploy script"
cat > /home/ec2-user/deploy.sh <<EOF
#!/bin/bash
set -e

echo "Downloading latest binary from S3..."
aws s3 cp s3://$S3_BUCKET/$BINARY_NAME $APP_DIR/$BINARY_NAME
chmod +x $APP_DIR/$BINARY_NAME

echo "Restarting Service"
sudo systemctl restart $SERVICE_NAME

echo "Deployed Successfully"
sudo systemctl status $SERVICE_NAME
EOF

chmod +x /home/ec2-user/deploy.sh
chown ec2-user:ec2-user /home/ec2-user/deploy.sh

echo "Setup Complete! Craftsbite is running..."