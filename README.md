# 编译命令
编译main.go和server.go生成server文件

go build -o server main.go server.go
# 运行server
./server 
# 模拟客户端
命令行输入

nc 127.0.0.1 8888