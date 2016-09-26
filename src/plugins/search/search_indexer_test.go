package search

import (
	"testing"

	"github.com/Dataman-Cloud/crane/src/utils/config"

	"github.com/Dataman-Cloud/crane/src/dockerclient"
	mock "github.com/Dataman-Cloud/crane/src/testing"
	"github.com/stretchr/testify/assert"
)

func TestNewCraneIndex(t *testing.T) {
	craneIndex := NewCraneIndex(&dockerclient.CraneDockerClient{})
	assert.Equal(t, craneIndex.CraneDockerClient, &dockerclient.CraneDockerClient{}, "should be equal")
}

func TestSearchIndex(t *testing.T) {
	mockServer := mock.NewServer()
	defer mockServer.Close()
	envs := map[string]interface{}{
		"Version":       "1.10.1",
		"Os":            "linux",
		"KernelVersion": "3.13.0-77-generic",
		"GoVersion":     "go1.4.2",
		"GitCommit":     "9e83765",
		"Arch":          "amd64",
		"ApiVersion":    "1.22",
		"BuildTime":     "2015-12-01T07:09:13.444803460+00:00",
		"Experimental":  false,
	}
	mockServer.AddRouter("/_ping", "get").
		Reply(200)
	mockServer.AddRouter("/version", "get").
		Reply(200).
		WJSON(envs)
	mockServer.AddRouter("/nodes", "get").
		Reply(200).
		WFile("./test/data/nodes.json")
	mockServer.Register()

	config := &config.Config{
		DockerEntryScheme: mockServer.Scheme,
		SwarmManagerIP:    mockServer.Addr,
		DockerEntryPort:   mockServer.Port,
		DockerTlsVerify:   false,
		DockerApiVersion:  "",
	}
	craneDockerClient, err := dockerclient.NewCraneDockerClient(config)
	if err != nil {
		t.Error("fails to create CraneDockerClient:", err)
	}
	craneIndexer := &CraneIndexer{
		CraneDockerClient: craneDockerClient,
	}
	documentStorage := &DocumentStorage{
		Store: map[string]Document{},
	}
	craneIndexer.Index(documentStorage)
}
