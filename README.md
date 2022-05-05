# 编译命令
编译main.go和server.go生成server文件

go build -o server main.go server.go

编译client.go生成client文件

go build -o client client.go
# 运行server和client
./server 

./client
# 模拟客户端
前期可使用命令行输入模拟客户端，后期完成client.go就不需要

nc 127.0.0.1 8888