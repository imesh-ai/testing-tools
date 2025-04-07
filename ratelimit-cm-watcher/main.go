// https://github.com/envoyproxy/ratelimit/tree/main/examples/xds-sotw-config-server
// package main

// import (
// 	"context"
// 	"flag"
// 	"os"

// 	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
// 	"github.com/envoyproxy/go-control-plane/pkg/server/v3"
// 	"github.com/envoyproxy/go-control-plane/pkg/test/v3"
// 	example "github.com/envoyproxy/ratelimit/examples/xds-sotw-config-server"
// 	envoyLogger "imesh.ai/ratelimit-cm-watcher/envoyLogger"
// 	"imesh.ai/ratelimit-cm-watcher/logger"
// )

// var (
// 	cacheLogger envoyLogger.Logger
// 	port        uint
// 	nodeID      string
// )

// func init() {
// 	// 	flag.UintVar(&port, "port", 18000, "xDS management server port")
// 	// 	flag.StringVar(&nodeID, "nodeID", "test-node-id", "Node ID")

// 	logger.InitLogger()

// 	port = 18000
// 	nodeID = "test-node-id"
// 	cacheLogger = envoyLogger.Logger{
// 		Debug: true,
// 	}
// }

// func main() {
// 	flag.Parse()

// 	// Create a cache
// 	cache := cache.NewSnapshotCache(false, cache.IDHash{}, cacheLogger)

// 	// Create the snapshot that we'll serve to Envoy
// 	snapshot := example.GenerateSnapshot()
// 	if err := snapshot.Consistent(); err != nil {
// 		logger.L.Sugar().Errorf("Snapshot is inconsistent: %+v\n%+v", snapshot, err)
// 		os.Exit(1)
// 	}
// 	logger.L.Sugar().Debugf("Will serve snapshot %+v", snapshot)

// 	// Add the snapshot to the cache
// 	if err := cache.SetSnapshot(context.Background(), nodeID, snapshot); err != nil {
// 		logger.L.Sugar().Errorf("Snapshot error %q for %+v", err, snapshot)
// 		os.Exit(1)
// 	}

// 	// Run the xDS server
// 	ctx := context.Background()
// 	cb := &test.Callbacks{Debug: cacheLogger.Debug}
// 	srv := server.NewServer(ctx, cache, cb)
// 	example.RunServer(ctx, srv, port)
// }

package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"go.uber.org/zap"
	"gopkg.in/yaml.v2"

	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"github.com/envoyproxy/go-control-plane/pkg/test/v3"
	rls_config "github.com/envoyproxy/go-control-plane/ratelimit/config/ratelimit/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"imesh.ai/ratelimit-cm-watcher/envoyLogger"
	"imesh.ai/ratelimit-cm-watcher/kubernetesClient"
	"imesh.ai/ratelimit-cm-watcher/logger"
	apiV1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	logger.InitLogger()
}

const (
	grpcKeepaliveTime        = 30 * time.Second
	grpcKeepaliveTimeout     = 5 * time.Second
	grpcKeepaliveMinTime     = 30 * time.Second
	grpcMaxConcurrentStreams = 1000000
)

// RunServer starts an xDS server at the given port.
func RunServer(ctx context.Context, srv server.Server, port uint) {
	grpcServer := grpc.NewServer(
		grpc.MaxConcurrentStreams(grpcMaxConcurrentStreams),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    grpcKeepaliveTime,
			Timeout: grpcKeepaliveTimeout,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             grpcKeepaliveMinTime,
			PermitWithoutStream: true,
		}),
	)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal(err)
	}

	discovery.RegisterAggregatedDiscoveryServiceServer(grpcServer, srv)

	log.Printf("Management server listening on %d\n", port)
	if err = grpcServer.Serve(lis); err != nil {
		log.Println(err)
	}
}

func main() {
	client := kubernetesClient.GetClient()

	ctx := context.Background()

	watcher, err := client.CoreV1().ConfigMaps("default").Watch(ctx, v1.ListOptions{})
	if err != nil {
		panic(err)
	}

	cacheLogger := envoyLogger.Logger{
		Debug: true,
	}
	snapshotCache := cache.NewSnapshotCache(
		false,
		cache.IDHash{},
		cacheLogger,
	)

	port := uint(18000)
	srv := server.NewServer(
		ctx,
		snapshotCache,
		&test.Callbacks{Debug: cacheLogger.Debug},
	)
	go RunServer(
		ctx,
		srv,
		port,
	)

	for event := range watcher.ResultChan() {
		cm := event.Object.(*apiV1.ConfigMap)

		log.Printf("%s %s/%s", event.Type, cm.ObjectMeta.Namespace, cm.ObjectMeta.Name)

		watchCM := false
		key := ""
		for k, v := range cm.ObjectMeta.Labels {
			if k == "ai.imesh.watch" && v == "true" {
				if len(cm.Data) != 1 {
					log.Printf("%s/%s does not have label", cm.ObjectMeta.Namespace, cm.ObjectMeta.Name)
					continue
				}
				watchCM = true

				for k, _ := range cm.Data {
					key = k
				}
				log.Printf("%s/%s does have required label and selected %s as key", cm.ObjectMeta.Namespace, cm.ObjectMeta.Name, key)

				break
			}
		}

		if !watchCM {
			log.Printf("Skipping %s/%s", cm.ObjectMeta.Namespace, cm.ObjectMeta.Name)
			continue
		}

		var config rls_config.RateLimitConfig

		err = yaml.Unmarshal([]byte(cm.Data[key]), &config)

		if err != nil {
			log.Print("Failed to parse", zap.Error(err))
			continue
		}
	}
}
