#!/bin/bash
set -e

# 创建日志文件
LOG_FILE="deploy.log"
echo "Starting deployment process..." > $LOG_FILE

# 清理旧容器和镜像
echo "Cleaning up old containers and images..." >> $LOG_FILE
docker rm -f duang-backend 2>/dev/null || true
docker rmi -f duang-backend 2>/dev/null || true

# 拉取基础镜像
echo "Pulling base images..." >> $LOG_FILE
docker pull golang:1.20 >> $LOG_FILE 2>&1

# 构建镜像
echo "Building duang-backend image..." >> $LOG_FILE
docker build -t duang-backend . >> $LOG_FILE 2>&1

# 检查端口是否被占用
echo "Checking if port 8081 is in use..." >> $LOG_FILE
if lsof -i :8081 >> $LOG_FILE 2>&1; then
  echo "Port 8081 is already in use, using port 8082 instead." >> $LOG_FILE
  PORT=8082
else
  echo "Port 8081 is available." >> $LOG_FILE
  PORT=8081
fi

# 运行容器
echo "Running duang-backend container on port $PORT..." >> $LOG_FILE
docker run -d -p $PORT:8081 --name duang-backend duang-backend >> $LOG_FILE 2>&1

# 检查容器状态
echo "Checking container status..." >> $LOG_FILE
docker ps | grep duang-backend >> $LOG_FILE 2>&1 || echo "Container not running!" >> $LOG_FILE

echo "Deployment completed. Check $LOG_FILE for details."
cat $LOG_FILE