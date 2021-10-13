# registry-zookeeper

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
	svr := echo.NewServer(new(EchoImpl), server.WithRegistry(zkregistry.NewZookeeperRegistry([]string{"127.0.0.1:2181"}, 40*time.Second)))
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
    client, err := echo.NewClient("echo", client.WithResolver(resolver.NewZookeeperResolver([]string{"127.0.0.1:2181"}, 40*time.Second)))
	if err != nil {
		log.Fatal(err)
	}
    ...
}
```

## More info

See discovery_test.go
