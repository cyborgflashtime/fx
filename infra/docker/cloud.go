package docker

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/metrue/fx/infra"
	"github.com/metrue/fx/types"
	"github.com/metrue/go-ssh-client"
	"github.com/mitchellh/go-homedir"
)

// Cloud define a docker host
type Cloud struct {
	IP   string `json:"ip"`
	User string `json:"user"`
	Name string `json:"name"`
	Type string `json:"type"`

	sshClient ssh.Clienter
}

// New new a docker cloud
func New(ip string, user string, name string) *Cloud {
	return &Cloud{
		IP:   ip,
		User: user,
		Name: name,
		Type: types.CloudTypeDocker,
	}
}

// Create a docker node
func Create(ip string, user string, name string) (*Cloud, error) {
	key, err := sshkey()
	if err != nil {
		return nil, err
	}
	port := sshport()
	sshClient := ssh.New(ip).WithUser(user).WithKey(key).WithPort(port)
	return &Cloud{
		IP:   ip,
		User: user,
		Name: name,
		Type: types.CloudTypeDocker,

		sshClient: sshClient,
	}, nil
}

// Load a docker node from meta
func Load(meta []byte) (*Cloud, error) {
	var cloud Cloud
	if err := json.Unmarshal(meta, &cloud); err != nil {
		return nil, err
	}
	key, err := sshkey()
	if err != nil {
		return nil, err
	}
	port := sshport()
	sshClient := ssh.New(cloud.IP).WithUser(cloud.User).WithKey(key).WithPort(port)
	cloud.sshClient = sshClient

	return &cloud, nil
}

// Provision a host
func (c *Cloud) Provision() error {
	if err := c.sshClient.RunCommand(infra.Scripts["docker_version"].(string), ssh.CommandOptions{}); err != nil {
		if err := c.sshClient.RunCommand(infra.Scripts["install_docker"].(string), ssh.CommandOptions{}); err != nil {
			return err
		}

		if err := c.sshClient.RunCommand(infra.Scripts["start_dockerd"].(string), ssh.CommandOptions{}); err != nil {
			return err
		}
	}

	if err := c.sshClient.RunCommand(infra.Scripts["check_fx_agent"].(string), ssh.CommandOptions{}); err != nil {
		if err := c.sshClient.RunCommand(infra.Scripts["start_fx_agent"].(string), ssh.CommandOptions{}); err != nil {
			return err
		}
	}
	return nil
}

// GetType cloud type
func (c *Cloud) GetType() string {
	return c.Type
}

func (c *Cloud) GetConfig() (string, error) {
	data, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (c *Cloud) Dump() ([]byte, error) {
	return json.Marshal(c)
}

// NOTE only using for unit testing
func (c *Cloud) setsshClient(client ssh.Clienter) {
	c.sshClient = client
}

// NOTE the reason putting sshkey() and sshport here inside node.go is because
// ssh key and ssh port is related to node it self, we may extend this in future
func sshkey() (string, error) {
	path := os.Getenv("SSH_KEY_FILE")
	if path != "" {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return "", err
		}
		return absPath, nil
	}

	key, err := homedir.Expand("~/.ssh/id_rsa")
	if err != nil {
		return "", err
	}
	return key, nil
}

func sshport() string {
	port := os.Getenv("SSH_PORT")
	if port != "" {
		return port
	}
	return "22"
}

var (
	_ infra.Clouder = &Cloud{}
)