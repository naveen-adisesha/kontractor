package contract

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func LoadFromFile(path string) (*ContainerContract, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading contract file %s: %w", path, err)
	}

	var c ContainerContract
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("parsing contract YAML: %w", err)
	}

	if c.APIVersion != "kontractor.io/v1" {
		return nil, fmt.Errorf("unsupported apiVersion %q (expected kontractor.io/v1)", c.APIVersion)
	}
	if c.Kind != "ContainerContract" {
		return nil, fmt.Errorf("unsupported kind %q (expected ContainerContract)", c.Kind)
	}

	return &c, nil
}
