package main

import (
	"fmt"

	"web-server-eccho/infra/k8s"

	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(deploy)
}

func deploy(ctx *pulumi.Context) error {
	cfg := config.New(ctx, "")
	image := cfg.Get("image")
	if image == "" {
		image = "web-server-eccho:local"
	}
	cluster := cfg.Get("clusterName")
	if cluster == "" {
		cluster = "echo"
	}
	port := 8080
	if cfg.Get("port") != "" {
		port = cfg.GetInt("port")
	}
	if port < 1 || port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}

	provider, err := kubernetes.NewProvider(ctx, "kind", &kubernetes.ProviderArgs{
		Context: pulumi.String("kind-" + cluster),
	})
	if err != nil {
		return err
	}
	ns, err := corev1.NewNamespace(ctx, "echo", &corev1.NamespaceArgs{
		Metadata: &metav1.ObjectMetaArgs{Name: pulumi.String("echo")},
	}, pulumi.Provider(provider))
	if err != nil {
		return err
	}
	labels := pulumi.StringMap{"app": pulumi.String("echo")}
	deployment, err := k8s.NewDeployment(ctx, ns.Metadata.Name(), labels, image, port, pulumi.Provider(provider))
	if err != nil {
		return err
	}
	service, err := k8s.NewService(ctx, ns.Metadata.Name(), labels, pulumi.Provider(provider))
	if err != nil {
		return err
	}
	ctx.Export("namespace", ns.Metadata.Name())
	ctx.Export("deployment", deployment.Metadata.Name())
	ctx.Export("service", service.Metadata.Name())
	return nil
}
