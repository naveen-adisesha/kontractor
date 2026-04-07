package parser

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"

	"gopkg.in/yaml.v3"
)

type K8sResource struct {
	APIVersion string                 `yaml:"apiVersion"`
	Kind       string                 `yaml:"kind"`
	Metadata   ResourceMetadata       `yaml:"metadata"`
	Spec       map[string]interface{} `yaml:"spec,omitempty"`
	Data       map[string]interface{} `yaml:"data,omitempty"`
	RawNode    *yaml.Node
	RawBytes   []byte
}

type ResourceMetadata struct {
	Name        string            `yaml:"name"`
	Namespace   string            `yaml:"namespace"`
	Labels      map[string]string `yaml:"labels"`
	Annotations map[string]string `yaml:"annotations"`
}

func IsWorkload(kind string) bool {
	switch kind {
	case "Deployment", "StatefulSet", "DaemonSet", "Job", "CronJob":
		return true
	}
	return false
}

func ParseManifests(r io.Reader) ([]*yaml.Node, [][]byte, error) {
	var nodes []*yaml.Node
	var rawChunks [][]byte

	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024)

	var current bytes.Buffer
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "---" {
			if current.Len() > 0 {
				chunk := make([]byte, current.Len())
				copy(chunk, current.Bytes())
				node, err := parseNode(chunk)
				if err == nil && node != nil {
					nodes = append(nodes, node)
					rawChunks = append(rawChunks, chunk)
				}
			}
			current.Reset()
			continue
		}
		current.WriteString(line + "\n")
	}

	if current.Len() > 0 {
		chunk := make([]byte, current.Len())
		copy(chunk, current.Bytes())
		node, err := parseNode(chunk)
		if err == nil && node != nil {
			nodes = append(nodes, node)
			rawChunks = append(rawChunks, chunk)
		}
	}

	return nodes, rawChunks, scanner.Err()
}

func parseNode(data []byte) (*yaml.Node, error) {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return nil, nil
	}

	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, err
	}

	if doc.Kind == yaml.DocumentNode && len(doc.Content) > 0 {
		return doc.Content[0], nil
	}
	return &doc, nil
}

func GetField(node *yaml.Node, path ...string) *yaml.Node {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i < len(node.Content)-1; i += 2 {
		if node.Content[i].Value == path[0] {
			if len(path) == 1 {
				return node.Content[i+1]
			}
			return GetField(node.Content[i+1], path[1:]...)
		}
	}
	return nil
}

func GetStringField(node *yaml.Node, path ...string) string {
	n := GetField(node, path...)
	if n == nil {
		return ""
	}
	return n.Value
}

func SerializeManifests(nodes []*yaml.Node) ([]byte, error) {
	var buf bytes.Buffer
	for i, node := range nodes {
		if i > 0 {
			buf.WriteString("---\n")
		}

		doc := &yaml.Node{
			Kind:    yaml.DocumentNode,
			Content: []*yaml.Node{node},
		}
		enc := yaml.NewEncoder(&buf)
		enc.SetIndent(2)
		if err := enc.Encode(doc); err != nil {
			return nil, fmt.Errorf("serializing resource %d: %w", i, err)
		}
		enc.Close()
	}
	return buf.Bytes(), nil
}
