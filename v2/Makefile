.PHONY: setup-rabbitmq build run-dataserver clean

# 设置RabbitMQ交换机
setup-rabbitmq:
	@echo "Setting up RabbitMQ exchanges..."
	cd scripts && go run setup-rabbitmq.go

# 构建数据服务器
build:
	@echo "Building dataserver..."
	cd dataserver/main && go build -o ../../bin/dataserver .

# 运行数据服务器（需要先设置RabbitMQ）
run-dataserver: setup-rabbitmq build
	@echo "Starting dataserver..."
	./bin/dataserver

# 清理构建文件
clean:
	rm -rf bin/

# 开发环境快速启动
dev: setup-rabbitmq
	@echo "Starting dataserver in development mode..."
	cd dataserver/main && go run .

# 检查RabbitMQ连接
check-rabbitmq:
	@echo "Checking RabbitMQ connection..."
	cd scripts && go run setup-rabbitmq.go
