package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type ContainerContract struct {
	APIVersion string   `yaml:"apiVersion"`
	Kind       string   `yaml:"kind"`
	Metadata   Metadata `yaml:"metadata"`
	Spec       Spec     `yaml:"spec"`
}

type Metadata struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Image       string `yaml:"image"`
	Description string `yaml:"description"`
}

type Spec struct {
	Features map[string]Feature `yaml:"features"`
	Env      EnvSpec            `yaml:"env"`
	Mounts   MountSpec          `yaml:"mounts"`
	Ports    []Port             `yaml:"ports"`
	Health   HealthSpec         `yaml:"health"`
	Security SecuritySpec       `yaml:"security"`
	RBAC     RBACSpec           `yaml:"rbac"`
}

type Feature struct {
	Description string `yaml:"description"`
	Default     bool   `yaml:"default"`
}

type EnvSpec struct {
	Required    []EnvVar             `yaml:"required"`
	Optional    []EnvVar             `yaml:"optional"`
	Conditional []ConditionalEnvGroup `yaml:"conditional"`
}

type EnvVar struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Default     string `yaml:"default"`
	Value       string `yaml:"value"`
}

type ConditionalEnvGroup struct {
	When WhenClause `yaml:"when"`
	Env  []EnvVar   `yaml:"env"`
}

type WhenClause struct {
	Feature string `yaml:"feature"`
}

type MountSpec struct {
	Required    []Mount               `yaml:"required"`
	Optional    []Mount               `yaml:"optional"`
	Conditional []ConditionalMountGroup `yaml:"conditional"`
}

type Mount struct {
	Path        string `yaml:"path"`
	Type        string `yaml:"type"`
	MaxSize     string `yaml:"maxSize"`
	MinSize     string `yaml:"minSize"`
	SecretName  string `yaml:"secretName"`
	Description string `yaml:"description"`
}

type ConditionalMountGroup struct {
	When   WhenClause `yaml:"when"`
	Mounts []Mount    `yaml:"mounts"`
}

type Port struct {
	Name          string `yaml:"name"`
	ContainerPort int    `yaml:"containerPort"`
	Protocol      string `yaml:"protocol"`
	Description   string `yaml:"description"`
}

type HealthSpec struct {
	Startup   *Probe `yaml:"startup"`
	Readiness *Probe `yaml:"readiness"`
	Liveness  *Probe `yaml:"liveness"`
}

type Probe struct {
	HTTPGet             *HTTPGetAction `yaml:"httpGet"`
	InitialDelaySeconds int            `yaml:"initialDelaySeconds"`
	PeriodSeconds       int            `yaml:"periodSeconds"`
}

type HTTPGetAction struct {
	Path   string `yaml:"path"`
	Port   int    `yaml:"port"`
	Scheme string `yaml:"scheme"`
}

type SecuritySpec struct {
	RunAsNonRoot           bool     `yaml:"runAsNonRoot"`
	ReadOnlyRootFilesystem bool     `yaml:"readOnlyRootFilesystem"`
	DropCapabilities       []string `yaml:"dropCapabilities"`
	PreferredUID           int      `yaml:"preferredUid"`
	PreferredGID           int      `yaml:"preferredGid"`
}

type RBACSpec struct {
	Rules        []RBACRule `yaml:"rules"`
	ClusterRules []RBACRule `yaml:"clusterRules"`
}

type RBACRule struct {
	APIGroups []string `yaml:"apiGroups"`
	Resources []string `yaml:"resources"`
	Verbs     []string `yaml:"verbs"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: kontractor-validate <contract.yaml> [--features feature1,feature2]\n")
		os.Exit(1)
	}

	contractPath := os.Args[1]
	data, err := os.ReadFile(contractPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading contract: %v\n", err)
		os.Exit(1)
	}

	var contract ContainerContract
	if err := yaml.Unmarshal(data, &contract); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing contract: %v\n", err)
		os.Exit(1)
	}

	if contract.APIVersion != "kontractor.io/v1" {
		fmt.Fprintf(os.Stderr, "Unsupported apiVersion: %s (expected kontractor.io/v1)\n", contract.APIVersion)
		os.Exit(1)
	}

	fmt.Printf("Contract: %s v%s\n", contract.Metadata.Name, contract.Metadata.Version)
	fmt.Printf("Image:    %s\n", contract.Metadata.Image)
	fmt.Printf("Description: %s\n\n", contract.Metadata.Description)

	fmt.Println("=== Features ===")
	for name, f := range contract.Spec.Features {
		fmt.Printf("  [%s] %s (default: %v)\n", name, f.Description, f.Default)
	}

	fmt.Println("\n=== Required Environment Variables ===")
	for _, e := range contract.Spec.Env.Required {
		def := e.Default
		if def == "" {
			def = "(no default)"
		}
		fmt.Printf("  %s = %s\n", e.Name, def)
	}

	fmt.Println("\n=== Required Mounts ===")
	for _, m := range contract.Spec.Mounts.Required {
		fmt.Printf("  %s [%s] %s\n", m.Path, m.Type, m.MaxSize)
	}

	for _, cg := range contract.Spec.Mounts.Conditional {
		fmt.Printf("\n=== Conditional Mounts (when: %s) ===\n", cg.When.Feature)
		for _, m := range cg.Mounts {
			fmt.Printf("  %s [%s] %s\n", m.Path, m.Type, m.Description)
		}
	}

	fmt.Println("\n=== Ports ===")
	for _, p := range contract.Spec.Ports {
		fmt.Printf("  %s: %d/%s - %s\n", p.Name, p.ContainerPort, p.Protocol, p.Description)
	}

	fmt.Println("\n=== RBAC Rules ===")
	for _, r := range contract.Spec.RBAC.Rules {
		fmt.Printf("  %v %v %v\n", r.APIGroups, r.Resources, r.Verbs)
	}

	fmt.Println("\n=== Security ===")
	fmt.Printf("  runAsNonRoot: %v\n", contract.Spec.Security.RunAsNonRoot)
	fmt.Printf("  readOnlyRootFilesystem: %v\n", contract.Spec.Security.ReadOnlyRootFilesystem)
	fmt.Printf("  preferredUid: %d\n", contract.Spec.Security.PreferredUID)
	fmt.Printf("  dropCapabilities: %v\n", contract.Spec.Security.DropCapabilities)

	fmt.Println("\n[PASS] Contract is valid.")
}
