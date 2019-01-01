### 运行

    docker run -d -p 9411:9411 openzipkin/zipkin

### 编译并运行grpc-server
    cd grpc-server
    go build server1.go
    go build server2.go
    go build server3.go
    ./server1
    ./server2
    ./server3

### 编译并运行grpc-client
    cd grpc-client
    go build client.go
    ./client


### 查看调用链分析

浏览器中打开http://127.0.0.1:9411/zipkin/