# registry-zookeeper (*This is a community driven project*)

Some applications use Zookeeper as service discovery.

## How to use?

### Server
```go
import (
    ...
	zkregistry "github.com/kitex-contrib/registry-zookeeper/registry"
	"github.com/cloudwego/kitex/server"
    ...
)

func main() {
    ...
    r, err := zkregistry.NewZookeeperRegistry([]string{"127.0.0.1:2181"}, 40*time.Second)
    if err != nil{
        panic(err)
    }
    svr := echo.NewServer(new(EchoImpl), server.WithRegistry(r))
    if err := svr.Run(); err != nil {
    log.Println("server stopped with error:", err)
    } else {
        log.Println("server stopped")
    }
    ...
}


```

### Client
```go
import (
    ...
    "github.com/kitex-contrib/registry-zookeeper/resolver"
    "github.com/cloudwego/kitex/client"
    ...
)

func main() {
    ...
    r, err := resolver.NewZookeeperResolver([]string{"127.0.0.1:2181"}, 40*time.Second)
    if err != nil {
        panic(err)
    }
    client, err := echo.NewClient("echo", client.WithResolver(r))
	if err != nil {
		log.Fatal(err)
	}
    ...
}
```

## Authentication

### Server
```go
import (
    ...
	zkregistry "github.com/kitex-contrib/registry-zookeeper/registry"
	"github.com/cloudwego/kitex/server"
    ...
)

func main() {
    ...
    // creates a zk based registry with given username and password.
    r, err := zkregistry.NewZookeeperRegistryWithAuth([]string{"127.0.0.1:2181"}, 40*time.Second, "username", "password")
    if err != nil{
        panic(err)
    }
    svr := echo.NewServer(new(EchoImpl), server.WithRegistry(r))
    if err := svr.Run(); err != nil {
       log.Println("server stopped with error:", err)
    } else {
        log.Println("server stopped")
    }
    ...
}
```

### Client
```go
import (
    ...
    "github.com/kitex-contrib/registry-zookeeper/resolver"
    "github.com/cloudwego/kitex/client"
    ...
)

func main() {
    ...
	// creates a zk based resolver with given username and password.
    r, err := resolver.NewZookeeperResolverWithAuth([]string{"127.0.0.1:2181"}, 40*time.Second, "username", "password")
    if err != nil {
        panic(err)
    }
    client, err := echo.NewClient("echo", client.WithResolver(r))
	if err != nil {
		log.Fatal(err)
	}
    ...
}
```

## More info

See discovery_test.go


## Compatibility

Compatible with server (3.4.0 - 3.7.0), If you want to use older server version, please modify the version in `Makefile` to test. 

[zookeeper server version list]( https://zookeeper.apache.org/documentation.html)




maintained by: horizonzy (horizonzy@apache.org)