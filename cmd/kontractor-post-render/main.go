package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/kontractor/kontractor/pkg/contract"
	"github.com/kontractor/kontractor/pkg/mutator"
	"github.com/kontractor/kontractor/pkg/parser"
	"gopkg.in/yaml.v3"
)

func main() {
	contractFile := flag.String("contract", "", "Path to container contract YAML file")
	featuresFlag := flag.String("features", "", "Comma-separated feature flags to enable (e.g. tls,metrics)")
	setVars := flag.String("set-vars", "", "Variable substitutions in KEY=VALUE,... format (e.g. RELEASE=myrelease)")
	dryRun := flag.Bool("dry-run", false, "Show what would be mutated without changing output")
	quiet := flag.Bool("quiet", false, "Suppress mutation report on stderr")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `kontractor-post-render - Helm post-renderer for Kontractor contracts

Usage as Helm post-renderer:
  helm install myrelease ./mychart \
    --post-renderer ./kontractor-post-render \
    --post-renderer-args="--contract=contract.yaml --features=tls"

Usage standalone:
  helm template ./mychart | ./kontractor-post-render --contract=contract.yaml

Flags:
`)
		flag.PrintDefaults()
	}

	flag.Parse()

	if *contractFile == "" {
		contractPath := os.Getenv("KONTRACTOR_CONTRACT")
		if contractPath != "" {
			contractFile = &contractPath
		} else {
			fmt.Fprintln(os.Stderr, "Error: --contract flag or KONTRACTOR_CONTRACT env var required")
			flag.Usage()
			os.Exit(1)
		}
	}

	cc, err := contract.LoadFromFile(*contractFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading contract: %v\n", err)
		os.Exit(1)
	}

	enabledFeatures := make(map[string]bool)
	if envFeatures := os.Getenv("KONTRACTOR_FEATURES"); envFeatures != "" {
		for _, f := range strings.Split(envFeatures, ",") {
			f = strings.TrimSpace(f)
			if f != "" {
				enabledFeatures[f] = true
			}
		}
	}
	if *featuresFlag != "" {
		for _, f := range strings.Split(*featuresFlag, ",") {
			f = strings.TrimSpace(f)
			if f != "" {
				enabledFeatures[f] = true
			}
		}
	}

	vars := make(map[string]string)
	if *setVars != "" {
		for _, kv := range strings.Split(*setVars, ",") {
			parts := strings.SplitN(kv, "=", 2)
			if len(parts) == 2 {
				vars[parts[0]] = parts[1]
			}
		}
	}

	nodes, _, err := parser.ParseManifests(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing stdin: %v\n", err)
		os.Exit(1)
	}

	if len(nodes) == 0 {
		fmt.Fprintln(os.Stderr, "Warning: no YAML documents found on stdin")
		os.Exit(0)
	}

	if _, ok := vars["RELEASE"]; !ok {
		if relName := detectReleaseName(nodes); relName != "" {
			vars["RELEASE"] = relName
		}
	}

	contract.SubstituteVars(cc, vars)

	resolvedFeatures := cc.ResolveFeatures(enabledFeatures)

	var results []*mutator.MutationResult
	for _, node := range nodes {
		kind := parser.GetStringField(node, "kind")
		if !parser.IsWorkload(kind) {
			continue
		}

		result, err := mutator.MutateWorkload(node, cc, resolvedFeatures)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error mutating %s: %v\n", kind, err)
			os.Exit(1)
		}
		results = append(results, result)
	}

	if *dryRun {
		if !*quiet {
			fmt.Fprint(os.Stderr, mutator.FormatResults(results))
			fmt.Fprintln(os.Stderr, "\n[dry-run] No output written")
		}
		os.Exit(0)
	}

	output, err := parser.SerializeManifests(nodes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error serializing output: %v\n", err)
		os.Exit(1)
	}

	os.Stdout.Write(output)

	if !*quiet {
		fmt.Fprint(os.Stderr, mutator.FormatResults(results))
	}
}

func detectReleaseName(nodes []*yaml.Node) string {
	for _, node := range nodes {
		name := parser.GetStringField(node, "metadata", "name")
		if name != "" {
			return name
		}
	}
	return ""
}
