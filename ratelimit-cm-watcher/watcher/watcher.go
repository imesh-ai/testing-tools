package watcher

import (
	"context"

	"imesh.ai/ratelimit-cm-watcher/kubernetesClient"
	"imesh.ai/ratelimit-cm-watcher/logger"
	apiV1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ListenForConfigMaps() {
	client := kubernetesClient.GetClient()

	ctx := context.Background()

	watcher, err := client.CoreV1().ConfigMaps("default").Watch(ctx, v1.ListOptions{})
	if err != nil {
		panic(err)
	}

	for event := range watcher.ResultChan() {
		cm := event.Object.(*apiV1.ConfigMap)

		logger.L.Sugar().Infof("%v", cm)
	}
}
