package main

import (
	"fmt"
	"sync"
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type resourceMocks struct {
	mu        sync.Mutex
	resources map[string]resource.PropertyMap
}

func (m *resourceMocks) NewResource(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
	outputs := args.Inputs.Copy()
	if metadata, ok := outputs["metadata"]; ok && metadata.IsObject() {
		values := metadata.ObjectValue().Copy()
		values["name"] = resource.NewStringProperty(args.Name)
		outputs["metadata"] = resource.NewObjectProperty(values)
	} else {
		outputs["metadata"] = resource.NewObjectProperty(resource.PropertyMap{"name": resource.NewStringProperty(args.Name)})
	}
	m.mu.Lock()
	m.resources[args.TypeToken] = outputs
	m.mu.Unlock()
	return args.Name + "-id", outputs, nil
}

func (*resourceMocks) Call(args pulumi.MockCallArgs) (resource.PropertyMap, error) {
	return args.Args, nil
}

func TestDeploymentWiring(t *testing.T) {
	mocks := &resourceMocks{resources: make(map[string]resource.PropertyMap)}
	err := pulumi.RunErr(deploy, pulumi.WithMocks("web-server-eccho", "test", mocks))
	if err != nil {
		t.Fatal(err)
	}
	provider := mocks.resources["pulumi:providers:kubernetes"]
	if got := provider["context"].StringValue(); got != "kind-echo" {
		t.Fatalf("context = %q", got)
	}
	deployment := mocks.resources["kubernetes:apps/v1:Deployment"]
	service := mocks.resources["kubernetes:core/v1:Service"]
	if deployment == nil || service == nil {
		t.Fatal("missing Deployment or Service")
	}
	spec := deployment["spec"].ObjectValue()
	template := spec["template"].ObjectValue()
	labels := template["metadata"].ObjectValue()["labels"].ObjectValue()
	serviceSpec := service["spec"].ObjectValue()
	selector := serviceSpec["selector"].ObjectValue()
	if labels["app"].StringValue() != selector["app"].StringValue() {
		t.Fatal("Service does not select the pod")
	}
	container := template["spec"].ObjectValue()["containers"].ArrayValue()[0].ObjectValue()
	if container["imagePullPolicy"].StringValue() != "Never" {
		t.Fatal("Kind must use its loaded image")
	}
	if container["image"].StringValue() != "web-server-eccho:local" {
		t.Fatal("unexpected default image")
	}
	port := container["ports"].ArrayValue()[0].ObjectValue()
	target := serviceSpec["ports"].ArrayValue()[0].ObjectValue()["targetPort"].StringValue()
	if target != port["name"].StringValue() {
		t.Fatal("Service target port does not match the container")
	}
	env := container["env"].ArrayValue()[0].ObjectValue()["value"].StringValue()
	if env != fmt.Sprintf("%.0f", port["containerPort"].NumberValue()) {
		t.Fatal("PORT does not match the container port")
	}
}
