package k8s

import (
	"context"
	"os"
	"testing"

	"github.com/metrue/fx/packer"
	"github.com/metrue/fx/types"
)

func TestK8SDeployer(t *testing.T) {
	name := "hellohello"
	bindings := []types.PortBinding{
		types.PortBinding{
			ServiceBindingPort:  80,
			ContainerExposePort: 3000,
		},
		types.PortBinding{
			ServiceBindingPort:  443,
			ContainerExposePort: 3000,
		},
	}
	kubeconfig := os.Getenv("KUBECONFIG")
	username := os.Getenv("DOCKER_USERNAME")
	password := os.Getenv("DOCKER_PASSWORD")
	if kubeconfig == "" || username == "" || password == "" {
		t.Skip("skip test since no KUBECONFIG, DOCKER_USERNAME and DOCKER_PASSWORD given in environment variable")
	}
	k8s, err := Create("")
	if err != nil {
		t.Fatal(err)
	}

	data, err := packer.PackIntoK8SConfigMapFile("./fixture")
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := k8s.Deploy(ctx, data, name, name, bindings); err != nil {
		t.Fatal(err)
	}

	if err := k8s.Destroy(ctx, name); err != nil {
		t.Fatal(err)
	}
}
