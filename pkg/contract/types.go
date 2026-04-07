package contract

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
	Features       map[string]Feature      `yaml:"features"`
	Env            EnvSpec                  `yaml:"env"`
	Mounts         MountSpec               `yaml:"mounts"`
	Secrets        SecretSpec              `yaml:"secrets"`
	Ports          []Port                   `yaml:"ports"`
	Health         HealthSpec               `yaml:"health"`
	Security       SecuritySpec             `yaml:"security"`
	RBAC           RBACSpec                 `yaml:"rbac"`
	InitContainers InitContainerSpec        `yaml:"initContainers"`
}

type Feature struct {
	Description string `yaml:"description"`
	Default     bool   `yaml:"default"`
}

type EnvSpec struct {
	Required    []EnvVar              `yaml:"required"`
	Optional    []EnvVar              `yaml:"optional"`
	Conditional []ConditionalEnvGroup `yaml:"conditional"`
}

type EnvVar struct {
	Name        string     `yaml:"name"`
	Description string     `yaml:"description"`
	Default     string     `yaml:"default"`
	Value       string     `yaml:"value"`
	Example     string     `yaml:"example"`
	SecretRef   *SecretRef `yaml:"secretRef,omitempty"`
}

type SecretRef struct {
	Name string `yaml:"name"`
	Key  string `yaml:"key"`
}

type ConditionalEnvGroup struct {
	When WhenClause `yaml:"when"`
	Env  []EnvVar   `yaml:"env"`
}

type WhenClause struct {
	Feature string `yaml:"feature"`
}

type MountSpec struct {
	Required    []Mount                 `yaml:"required"`
	Optional    []Mount                 `yaml:"optional"`
	Conditional []ConditionalMountGroup `yaml:"conditional"`
}

type Mount struct {
	Path          string   `yaml:"path"`
	Type          string   `yaml:"type"`
	MaxSize       string   `yaml:"maxSize"`
	MinSize       string   `yaml:"minSize"`
	AccessMode    string   `yaml:"accessMode"`
	SecretName    string   `yaml:"secretName"`
	ConfigMapName string   `yaml:"configMapName"`
	Keys          []string `yaml:"keys"`
	SubPath       string   `yaml:"subPath"`
	ReadOnly      bool     `yaml:"readOnly"`
	Description   string   `yaml:"description"`
}

type ConditionalMountGroup struct {
	When   WhenClause `yaml:"when"`
	Mounts []Mount    `yaml:"mounts"`
}

type SecretSpec struct {
	Required    []SecretRequirement              `yaml:"required"`
	Conditional []ConditionalSecretGroup         `yaml:"conditional"`
}

type SecretRequirement struct {
	Name        string   `yaml:"name"`
	Keys        []string `yaml:"keys"`
	MountPath   string   `yaml:"mountPath"`
	Description string   `yaml:"description"`
}

type ConditionalSecretGroup struct {
	When    WhenClause          `yaml:"when"`
	Secrets []SecretRequirement `yaml:"secrets"`
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
	HTTPGet             *HTTPGetAction         `yaml:"httpGet,omitempty"`
	TCPSocket           *TCPSocketAction       `yaml:"tcpSocket,omitempty"`
	Exec                *ExecAction            `yaml:"exec,omitempty"`
	InitialDelaySeconds int                    `yaml:"initialDelaySeconds"`
	PeriodSeconds       int                    `yaml:"periodSeconds"`
	TimeoutSeconds      int                    `yaml:"timeoutSeconds"`
	FailureThreshold    int                    `yaml:"failureThreshold"`
	SuccessThreshold    int                    `yaml:"successThreshold"`
	When                *ConditionalOverride   `yaml:"when,omitempty"`
}

type HTTPGetAction struct {
	Path   string `yaml:"path"`
	Port   int    `yaml:"port"`
	Scheme string `yaml:"scheme"`
}

type TCPSocketAction struct {
	Port int `yaml:"port"`
}

type ExecAction struct {
	Command []string `yaml:"command"`
}

type ConditionalOverride struct {
	Feature  string                 `yaml:"feature"`
	Override map[string]interface{} `yaml:"override"`
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
	APIGroups       []string `yaml:"apiGroups"`
	Resources       []string `yaml:"resources"`
	Verbs           []string `yaml:"verbs"`
	NonResourceURLs []string `yaml:"nonResourceURLs"`
}

type InitContainerSpec struct {
	Required    []InitSpec              `yaml:"required"`
	Conditional []ConditionalInitGroup  `yaml:"conditional"`
}

type InitSpec struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Purpose     string `yaml:"purpose"`
}

type ConditionalInitGroup struct {
	When           WhenClause `yaml:"when"`
	InitContainers []InitSpec `yaml:"initContainers"`
}

func (c *ContainerContract) ResolveFeatures(enabled map[string]bool) map[string]bool {
	resolved := make(map[string]bool)
	for name, f := range c.Spec.Features {
		if val, ok := enabled[name]; ok {
			resolved[name] = val
		} else {
			resolved[name] = f.Default
		}
	}
	return resolved
}

func (c *ContainerContract) ResolvedEnv(features map[string]bool) []EnvVar {
	var result []EnvVar
	result = append(result, c.Spec.Env.Required...)
	for _, cg := range c.Spec.Env.Conditional {
		if features[cg.When.Feature] {
			result = append(result, cg.Env...)
		}
	}
	return result
}

func (c *ContainerContract) ResolvedMounts(features map[string]bool) []Mount {
	var result []Mount
	result = append(result, c.Spec.Mounts.Required...)
	for _, cg := range c.Spec.Mounts.Conditional {
		if features[cg.When.Feature] {
			result = append(result, cg.Mounts...)
		}
	}
	return result
}
