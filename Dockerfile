FROM golang:1.20

WORKDIR /app

# 复制go.mod和go.sum文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download && go mod verify

# 复制源代码
COPY . .

# 打印Go版本和环境信息
RUN go version && go env

# 编译
RUN go build -v -o duang-backend main.go

EXPOSE 8081

CMD ["./duang-backend"]