package secrets

import (
	"encoding/json"
	"fmt"
	"github.com/gruntwork-io/gruntwork-cli/files"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

// Represents a typical web services or data store client that could be used to do CRUD. This one writes and reads from
// disk.
type MockClient struct {
	Username   string
	Password   string
	WorkingDir string
}

func NewClient(username string, password string) (*MockClient, error) {
	if username == "" || password == "" {
		return nil, fmt.Errorf("You must specify a username and password for the secrets provider")
	}

	workingDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	return &MockClient{Username: username, Password: password, WorkingDir: workingDir}, nil
}

func (client *MockClient) CreateSecret(name string, value string) (*Secret, error) {
	path := client.filePathForSecret(name)

	if files.FileExists(path) {
		return nil, fmt.Errorf("A secret named '%s' already exists", name)
	}

	log.Printf("[DEBUG] Writing secret to '%s'\n", path)

	secret := Secret{
		Name:  name,
		Value: value,
		Id:    path,
	}

	if err := writeSecret(secret); err != nil {
		return nil, err
	}

	return &secret, nil
}

func (client *MockClient) UpdateSecret(name string, value string) (*Secret, error) {
	path := client.filePathForSecret(name)

	log.Printf("[DEBUG] Updating secret at '%s'\n", path)

	secret := Secret{
		Name:  name,
		Value: value,
		Id:    path,
	}

	if err := writeSecret(secret); err != nil {
		return nil, err
	}

	return &secret, nil
}

func (client *MockClient) GetSecretByName(name string) (*Secret, error) {
	path := client.filePathForSecret(name)
	return client.GetSecretById(path)
}

// The id of a secret in this implementation is the file path
func (client *MockClient) GetSecretById(id string) (*Secret, error) {
	bytes, err := ioutil.ReadFile(id)
	if err != nil {
		return nil, err
	}

	var secret Secret
	if err := json.Unmarshal(bytes, &secret); err != nil {
		return nil, err
	}

	return &secret, nil
}

// The id of a secret in this implementation is the file path
func (client *MockClient) DeleteSecretById(id string) error {
	log.Printf("[DEBUG] Deleting secret at '%s'\n", id)

	err := os.Remove(id)
	// Terraform providers are not supposed to return errors if a resource has already been deleted. In this case,
	// that's if the file doesn't exist, so we ignore that type of error, but handle all other types.
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

func (client *MockClient) filePathForSecret(name string) string {
	return filepath.Join(client.WorkingDir, fmt.Sprintf("secret-%s.json", name))
}

func writeSecret(secret Secret) error {
	bytes, err := json.Marshal(secret)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(secret.Id, bytes, 0644)
}

type Secret struct {
	Name  string
	Value string
	Id    string
}
