package k8s

import (
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// NewService exposes the echo pods on port 80 within the cluster.
func NewService(ctx *pulumi.Context, namespace pulumi.StringPtrInput, labels pulumi.StringMap, opts ...pulumi.ResourceOption) (*corev1.Service, error) {
	return corev1.NewService(ctx, "echo", &corev1.ServiceArgs{
		Metadata: &metav1.ObjectMetaArgs{Namespace: namespace},
		Spec: &corev1.ServiceSpecArgs{
			Type: pulumi.String("ClusterIP"), Selector: labels,
			Ports: corev1.ServicePortArray{&corev1.ServicePortArgs{Port: pulumi.Int(80), TargetPort: pulumi.String("http")}},
		},
	}, opts...)
}
