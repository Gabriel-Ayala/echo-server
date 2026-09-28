package k8s

import (
	appsv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apps/v1"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"strconv"
)

// NewDeployment creates the echo workload with health probes and resource limits.
func NewDeployment(ctx *pulumi.Context, namespace pulumi.StringPtrInput, labels pulumi.StringMap, image string, port int, opts ...pulumi.ResourceOption) (*appsv1.Deployment, error) {
	return appsv1.NewDeployment(ctx, "echo", &appsv1.DeploymentArgs{
		Metadata: &metav1.ObjectMetaArgs{Namespace: namespace},
		Spec: &appsv1.DeploymentSpecArgs{
			Replicas: pulumi.Int(1),
			Selector: &metav1.LabelSelectorArgs{MatchLabels: labels},
			Template: &corev1.PodTemplateSpecArgs{
				Metadata: &metav1.ObjectMetaArgs{Labels: labels},
				Spec: &corev1.PodSpecArgs{
					Containers: corev1.ContainerArray{&corev1.ContainerArgs{
						Name: pulumi.String("echo"), Image: pulumi.String(image),
						ImagePullPolicy: pulumi.String("Never"),
						Env:             corev1.EnvVarArray{&corev1.EnvVarArgs{Name: pulumi.String("PORT"), Value: pulumi.String(strconv.Itoa(port))}},
						Ports:           corev1.ContainerPortArray{&corev1.ContainerPortArgs{Name: pulumi.String("http"), ContainerPort: pulumi.Int(port)}},
						ReadinessProbe:  &corev1.ProbeArgs{HttpGet: &corev1.HTTPGetActionArgs{Path: pulumi.String("/"), Port: pulumi.Int(port)}},
						LivenessProbe:   &corev1.ProbeArgs{HttpGet: &corev1.HTTPGetActionArgs{Path: pulumi.String("/"), Port: pulumi.Int(port)}},
						Resources: &corev1.ResourceRequirementsArgs{
							Requests: pulumi.StringMap{"cpu": pulumi.String("10m"), "memory": pulumi.String("16Mi")},
							Limits:   pulumi.StringMap{"cpu": pulumi.String("100m"), "memory": pulumi.String("64Mi")},
						},
					}},
				},
			},
		},
	}, opts...)
}
