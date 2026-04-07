package main

import (
	"fmt"
	"os"

	"github.com/kontractor/kontractor/pkg/contract"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: kontractor-validate <contract.yaml>\n")
		os.Exit(1)
	}

	cc, err := contract.LoadFromFile(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Contract: %s v%s\n", cc.Metadata.Name, cc.Metadata.Version)
	fmt.Printf("Image:    %s\n", cc.Metadata.Image)
	fmt.Printf("Description: %s\n\n", cc.Metadata.Description)

	fmt.Println("=== Features ===")
	for name, f := range cc.Spec.Features {
		fmt.Printf("  [%s] %s (default: %v)\n", name, f.Description, f.Default)
	}

	fmt.Println("\n=== Required Environment Variables ===")
	for _, e := range cc.Spec.Env.Required {
		def := e.Default
		if def == "" {
			def = "(no default)"
		}
		fmt.Printf("  %s = %s\n", e.Name, def)
	}

	fmt.Println("\n=== Required Mounts ===")
	for _, m := range cc.Spec.Mounts.Required {
		fmt.Printf("  %s [%s] %s\n", m.Path, m.Type, m.MaxSize)
	}

	for _, cg := range cc.Spec.Mounts.Conditional {
		fmt.Printf("\n=== Conditional Mounts (when: %s) ===\n", cg.When.Feature)
		for _, m := range cg.Mounts {
			fmt.Printf("  %s [%s] %s\n", m.Path, m.Type, m.Description)
		}
	}

	fmt.Println("\n=== Ports ===")
	for _, p := range cc.Spec.Ports {
		fmt.Printf("  %s: %d/%s - %s\n", p.Name, p.ContainerPort, p.Protocol, p.Description)
	}

	fmt.Println("\n=== RBAC Rules ===")
	for _, r := range cc.Spec.RBAC.Rules {
		fmt.Printf("  %v %v %v\n", r.APIGroups, r.Resources, r.Verbs)
	}

	fmt.Println("\n=== Security ===")
	fmt.Printf("  runAsNonRoot: %v\n", cc.Spec.Security.RunAsNonRoot)
	fmt.Printf("  readOnlyRootFilesystem: %v\n", cc.Spec.Security.ReadOnlyRootFilesystem)
	fmt.Printf("  preferredUid: %d\n", cc.Spec.Security.PreferredUID)
	fmt.Printf("  dropCapabilities: %v\n", cc.Spec.Security.DropCapabilities)

	fmt.Println("\n[PASS] Contract is valid.")
}
