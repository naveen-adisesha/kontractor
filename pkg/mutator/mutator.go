package mutator

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kontractor/kontractor/pkg/contract"
	"github.com/kontractor/kontractor/pkg/parser"
	"gopkg.in/yaml.v3"
)

type MutationResult struct {
	ResourceName string
	ResourceKind string
	EnvAdded     []string
	VolumesAdded []string
	MountsAdded  []string
	Skipped      bool
	Reason       string
}

func MutateWorkload(node *yaml.Node, cc *contract.ContainerContract, features map[string]bool) (*MutationResult, error) {
	kind := parser.GetStringField(node, "kind")
	name := parser.GetStringField(node, "metadata", "name")

	result := &MutationResult{
		ResourceName: name,
		ResourceKind: kind,
	}

	if !parser.IsWorkload(kind) {
		result.Skipped = true
		result.Reason = fmt.Sprintf("kind %s is not a workload", kind)
		return result, nil
	}

	resolved := cc.ResolveFeatures(features)

	podSpec := findPodSpec(node, kind)
	if podSpec == nil {
		result.Skipped = true
		result.Reason = "could not locate podSpec"
		return result, nil
	}

	containers := parser.GetField(podSpec, "containers")
	if containers == nil || containers.Kind != yaml.SequenceNode || len(containers.Content) == 0 {
		result.Skipped = true
		result.Reason = "no containers found in podSpec"
		return result, nil
	}

	mainContainer := containers.Content[0]

	envVars := cc.ResolvedEnv(resolved)
	for _, ev := range envVars {
		if !hasEnvVar(mainContainer, ev.Name) {
			addEnvVar(mainContainer, ev)
			result.EnvAdded = append(result.EnvAdded, ev.Name)
		}
	}

	mounts := cc.ResolvedMounts(resolved)
	for _, m := range mounts {
		volName := volumeNameFromPath(m.Path)
		if !hasVolume(podSpec, volName) {
			addVolume(podSpec, m, volName)
			result.VolumesAdded = append(result.VolumesAdded, volName)
		}
		if !hasVolumeMount(mainContainer, m.Path) {
			addVolumeMount(mainContainer, m, volName)
			result.MountsAdded = append(result.MountsAdded, m.Path)
		}
	}

	addContractAnnotation(node, cc, resolved)

	return result, nil
}

func findPodSpec(node *yaml.Node, kind string) *yaml.Node {
	if kind == "CronJob" {
		return parser.GetField(node, "spec", "jobTemplate", "spec", "template", "spec")
	}
	return parser.GetField(node, "spec", "template", "spec")
}

func hasEnvVar(container *yaml.Node, name string) bool {
	env := parser.GetField(container, "env")
	if env == nil {
		return false
	}
	for _, item := range env.Content {
		n := parser.GetStringField(item, "name")
		if n == name {
			return true
		}
	}
	return false
}

func addEnvVar(container *yaml.Node, ev contract.EnvVar) {
	env := parser.GetField(container, "env")

	value := ev.Default
	if ev.Value != "" {
		value = ev.Value
	}

	var envNode *yaml.Node
	if ev.SecretRef != nil {
		envNode = &yaml.Node{
			Kind: yaml.MappingNode,
			Content: []*yaml.Node{
				{Kind: yaml.ScalarNode, Value: "name"},
				{Kind: yaml.ScalarNode, Value: ev.Name},
				{Kind: yaml.ScalarNode, Value: "valueFrom"},
				{Kind: yaml.MappingNode, Content: []*yaml.Node{
					{Kind: yaml.ScalarNode, Value: "secretKeyRef"},
					{Kind: yaml.MappingNode, Content: []*yaml.Node{
						{Kind: yaml.ScalarNode, Value: "name"},
						{Kind: yaml.ScalarNode, Value: ev.SecretRef.Name},
						{Kind: yaml.ScalarNode, Value: "key"},
						{Kind: yaml.ScalarNode, Value: ev.SecretRef.Key},
					}},
				}},
			},
		}
	} else {
		envNode = &yaml.Node{
			Kind: yaml.MappingNode,
			Content: []*yaml.Node{
				{Kind: yaml.ScalarNode, Value: "name"},
				{Kind: yaml.ScalarNode, Value: ev.Name},
				{Kind: yaml.ScalarNode, Value: "value"},
				{Kind: yaml.ScalarNode, Value: value, Tag: "!!str"},
			},
		}
	}

	if env == nil {
		env = &yaml.Node{Kind: yaml.SequenceNode}
		container.Content = append(container.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: "env"},
			env,
		)
	}
	env.Content = append(env.Content, envNode)
}

func hasVolume(podSpec *yaml.Node, name string) bool {
	volumes := parser.GetField(podSpec, "volumes")
	if volumes == nil {
		return false
	}
	for _, v := range volumes.Content {
		n := parser.GetStringField(v, "name")
		if n == name {
			return true
		}
	}
	return false
}

func addVolume(podSpec *yaml.Node, m contract.Mount, volName string) {
	volumes := parser.GetField(podSpec, "volumes")

	volNode := buildVolumeNode(m, volName)

	if volumes == nil {
		volumes = &yaml.Node{Kind: yaml.SequenceNode}
		podSpec.Content = append(podSpec.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: "volumes"},
			volumes,
		)
	}
	volumes.Content = append(volumes.Content, volNode)
}

func buildVolumeNode(m contract.Mount, volName string) *yaml.Node {
	switch m.Type {
	case "secret":
		secretName := m.SecretName
		if secretName == "" {
			secretName = volName
		}
		return &yaml.Node{
			Kind: yaml.MappingNode,
			Content: []*yaml.Node{
				{Kind: yaml.ScalarNode, Value: "name"},
				{Kind: yaml.ScalarNode, Value: volName},
				{Kind: yaml.ScalarNode, Value: "secret"},
				{Kind: yaml.MappingNode, Content: []*yaml.Node{
					{Kind: yaml.ScalarNode, Value: "secretName"},
					{Kind: yaml.ScalarNode, Value: secretName},
				}},
			},
		}
	case "configmap", "configMap":
		cmName := m.ConfigMapName
		if cmName == "" {
			cmName = volName
		}
		return &yaml.Node{
			Kind: yaml.MappingNode,
			Content: []*yaml.Node{
				{Kind: yaml.ScalarNode, Value: "name"},
				{Kind: yaml.ScalarNode, Value: volName},
				{Kind: yaml.ScalarNode, Value: "configMap"},
				{Kind: yaml.MappingNode, Content: []*yaml.Node{
					{Kind: yaml.ScalarNode, Value: "name"},
					{Kind: yaml.ScalarNode, Value: cmName},
				}},
			},
		}
	case "ephemeral", "emptyDir":
		children := []*yaml.Node{
			{Kind: yaml.ScalarNode, Value: "name"},
			{Kind: yaml.ScalarNode, Value: volName},
			{Kind: yaml.ScalarNode, Value: "emptyDir"},
		}
		emptyDirContent := []*yaml.Node{}
		if m.MaxSize != "" {
			emptyDirContent = append(emptyDirContent,
				&yaml.Node{Kind: yaml.ScalarNode, Value: "sizeLimit"},
				&yaml.Node{Kind: yaml.ScalarNode, Value: m.MaxSize},
			)
		}
		children = append(children, &yaml.Node{Kind: yaml.MappingNode, Content: emptyDirContent})
		return &yaml.Node{Kind: yaml.MappingNode, Content: children}
	default:
		return &yaml.Node{
			Kind: yaml.MappingNode,
			Content: []*yaml.Node{
				{Kind: yaml.ScalarNode, Value: "name"},
				{Kind: yaml.ScalarNode, Value: volName},
				{Kind: yaml.ScalarNode, Value: "emptyDir"},
				{Kind: yaml.MappingNode},
			},
		}
	}
}

func hasVolumeMount(container *yaml.Node, path string) bool {
	mounts := parser.GetField(container, "volumeMounts")
	if mounts == nil {
		return false
	}
	for _, vm := range mounts.Content {
		mp := parser.GetStringField(vm, "mountPath")
		if mp == path {
			return true
		}
	}
	return false
}

func addVolumeMount(container *yaml.Node, m contract.Mount, volName string) {
	mounts := parser.GetField(container, "volumeMounts")

	mountNode := &yaml.Node{
		Kind: yaml.MappingNode,
		Content: []*yaml.Node{
			{Kind: yaml.ScalarNode, Value: "name"},
			{Kind: yaml.ScalarNode, Value: volName},
			{Kind: yaml.ScalarNode, Value: "mountPath"},
			{Kind: yaml.ScalarNode, Value: m.Path},
		},
	}

	if m.ReadOnly {
		mountNode.Content = append(mountNode.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: "readOnly"},
			&yaml.Node{Kind: yaml.ScalarNode, Value: "true", Tag: "!!bool"},
		)
	}
	if m.SubPath != "" {
		mountNode.Content = append(mountNode.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: "subPath"},
			&yaml.Node{Kind: yaml.ScalarNode, Value: m.SubPath},
		)
	}

	if mounts == nil {
		mounts = &yaml.Node{Kind: yaml.SequenceNode}
		container.Content = append(container.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: "volumeMounts"},
			mounts,
		)
	}
	mounts.Content = append(mounts.Content, mountNode)
}

func volumeNameFromPath(path string) string {
	name := strings.TrimPrefix(path, "/")
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, "_", "-")
	name = strings.ReplaceAll(name, ".", "-")
	if len(name) > 63 {
		name = name[:63]
	}
	return name
}

func addContractAnnotation(node *yaml.Node, cc *contract.ContainerContract, features map[string]bool) {
	metadata := parser.GetField(node, "metadata")
	if metadata == nil {
		return
	}

	annotations := parser.GetField(metadata, "annotations")
	if annotations == nil {
		annotations = &yaml.Node{Kind: yaml.MappingNode}
		metadata.Content = append(metadata.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: "annotations"},
			annotations,
		)
	}

	var activeFeatures []string
	for f, enabled := range features {
		if enabled {
			activeFeatures = append(activeFeatures, f)
		}
	}

	annotations.Content = append(annotations.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: "kontractor.io/contract"},
		&yaml.Node{Kind: yaml.ScalarNode, Value: cc.Metadata.Name + ":" + cc.Metadata.Version},
		&yaml.Node{Kind: yaml.ScalarNode, Value: "kontractor.io/mutated"},
		&yaml.Node{Kind: yaml.ScalarNode, Value: "true", Tag: "!!str"},
	)

	if len(activeFeatures) > 0 {
		annotations.Content = append(annotations.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: "kontractor.io/features"},
			&yaml.Node{Kind: yaml.ScalarNode, Value: strings.Join(activeFeatures, ",")},
		)
	}
}

func FormatResults(results []*MutationResult) string {
	var sb strings.Builder
	sb.WriteString("\n=== Kontractor Post-Render Mutation Report ===\n\n")

	mutated := 0
	for _, r := range results {
		if r.Skipped {
			continue
		}
		mutated++
		sb.WriteString(fmt.Sprintf("  %s/%s:\n", r.ResourceKind, r.ResourceName))
		if len(r.EnvAdded) > 0 {
			sb.WriteString(fmt.Sprintf("    env injected:     %s\n", strings.Join(r.EnvAdded, ", ")))
		}
		if len(r.VolumesAdded) > 0 {
			sb.WriteString(fmt.Sprintf("    volumes added:    %s\n", strings.Join(r.VolumesAdded, ", ")))
		}
		if len(r.MountsAdded) > 0 {
			sb.WriteString(fmt.Sprintf("    mounts added:     %s\n", strings.Join(r.MountsAdded, ", ")))
		}
	}

	sb.WriteString(fmt.Sprintf("\n  Total resources mutated: %s\n", strconv.Itoa(mutated)))
	sb.WriteString("===============================================\n")
	return sb.String()
}
