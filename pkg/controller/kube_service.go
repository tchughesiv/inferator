package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	cachev1 "sigs.k8s.io/controller-runtime/pkg/cache"
	clientv1 "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type KubernetesPlatformService struct {
	client          clientv1.Client
	discoveryclient *discovery.DiscoveryClient
	cache           cachev1.Cache
	scheme          *runtime.Scheme
}

func GetInstance(mgr manager.Manager) KubernetesPlatformService {
	discoveryclient, err := discovery.NewDiscoveryClientForConfig(mgr.GetConfig())
	if err != nil {
		log.Error(err, "Error getting image client.")
		return KubernetesPlatformService{}
	}
	return KubernetesPlatformService{
		client:          mgr.GetClient(),
		discoveryclient: discoveryclient,
		cache:           mgr.GetCache(),
		scheme:          mgr.GetScheme(),
	}
}

func (service *KubernetesPlatformService) Create(ctx context.Context, obj runtime.Object) error {
	return service.client.Create(ctx, obj)
}

func (service *KubernetesPlatformService) Delete(ctx context.Context, obj runtime.Object, opts ...clientv1.DeleteOption) error {
	return service.client.Delete(ctx, obj, opts...)
}

func (service *KubernetesPlatformService) Get(ctx context.Context, key clientv1.ObjectKey, obj runtime.Object) error {
	return service.client.Get(ctx, key, obj)
}

func (service *KubernetesPlatformService) List(ctx context.Context, list runtime.Object, opts clientv1.ListOption) error {
	return service.client.List(ctx, list, opts)
}

func (service *KubernetesPlatformService) Update(ctx context.Context, obj runtime.Object) error {
	return service.client.Update(ctx, obj)
}

func (service *KubernetesPlatformService) GetCached(ctx context.Context, key clientv1.ObjectKey, obj runtime.Object) error {
	return service.cache.Get(ctx, key, obj)
}

func (service *KubernetesPlatformService) GetScheme() *runtime.Scheme {
	return service.scheme
}

func (service *KubernetesPlatformService) GetDiscoveryClient() *discovery.DiscoveryClient {
	return service.discoveryclient
}

func (service *KubernetesPlatformService) IsMockService() bool {
	return false
}
