package compose

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"isrvd/pkgs/docker"
	"isrvd/pkgs/swarm"
	"isrvd/server/config"
)

func TestSwarmComposePrePullsImage(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, "config.yml")
	if err := os.WriteFile(configPath, []byte("server:\n  rootDirectory: "+root+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ISRVD_CONFIG", configPath)
	t.Setenv("ISRVD_MODE", "server")
	if err := config.Init(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = config.Close() })

	const image = "registry.example.test/app:latest"
	const pullError = "image pulls disabled on this daemon"
	engine := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("API-Version", "1.51")
		_, path, _ := strings.Cut(strings.TrimPrefix(r.URL.Path, "/"), "/")
		switch {
		case r.URL.Path == "/_ping":
			return
		case path == "services/app":
			http.Error(w, `{"message":"service not found"}`, http.StatusNotFound)
			return
		case path == "images/create":
			http.Error(w, `{"message":"`+pullError+`"}`, http.StatusForbidden)
			return
		default:
			t.Errorf("unexpected Docker request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}))
	t.Cleanup(engine.Close)

	dockerClient, err := docker.NewDockerService(&docker.DockerConfig{
		Host: strings.Replace(engine.URL, "http://", "tcp://", 1), ContainerRoot: root,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = dockerClient.Close() })
	swarmClient := swarm.NewSwarmService(dockerClient.Client(), func(string) string { return "" })
	svc, err := NewService(dockerClient, swarmClient)
	if err != nil {
		t.Fatal(err)
	}
	content := "name: example\nservices:\n  app:\n    image: " + image + "\n"
	if _, err := svc.SwarmDeploy(t.Context(), DeployRequest{Content: content, ForcePull: true}); err == nil || !strings.Contains(err.Error(), pullError) {
		t.Fatalf("Swarm deploy should pre-pull the image, got %v", err)
	}
}
