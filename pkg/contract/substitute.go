package contract

import "strings"

func SubstituteVars(cc *ContainerContract, vars map[string]string) {
	for i := range cc.Spec.Env.Required {
		cc.Spec.Env.Required[i].Default = sub(cc.Spec.Env.Required[i].Default, vars)
		cc.Spec.Env.Required[i].Value = sub(cc.Spec.Env.Required[i].Value, vars)
	}
	for i := range cc.Spec.Env.Conditional {
		for j := range cc.Spec.Env.Conditional[i].Env {
			cc.Spec.Env.Conditional[i].Env[j].Default = sub(cc.Spec.Env.Conditional[i].Env[j].Default, vars)
			cc.Spec.Env.Conditional[i].Env[j].Value = sub(cc.Spec.Env.Conditional[i].Env[j].Value, vars)
			if cc.Spec.Env.Conditional[i].Env[j].SecretRef != nil {
				cc.Spec.Env.Conditional[i].Env[j].SecretRef.Name = sub(cc.Spec.Env.Conditional[i].Env[j].SecretRef.Name, vars)
			}
		}
	}

	for i := range cc.Spec.Mounts.Required {
		cc.Spec.Mounts.Required[i].SecretName = sub(cc.Spec.Mounts.Required[i].SecretName, vars)
		cc.Spec.Mounts.Required[i].ConfigMapName = sub(cc.Spec.Mounts.Required[i].ConfigMapName, vars)
	}
	for i := range cc.Spec.Mounts.Conditional {
		for j := range cc.Spec.Mounts.Conditional[i].Mounts {
			cc.Spec.Mounts.Conditional[i].Mounts[j].SecretName = sub(cc.Spec.Mounts.Conditional[i].Mounts[j].SecretName, vars)
			cc.Spec.Mounts.Conditional[i].Mounts[j].ConfigMapName = sub(cc.Spec.Mounts.Conditional[i].Mounts[j].ConfigMapName, vars)
		}
	}

	for i := range cc.Spec.Secrets.Required {
		cc.Spec.Secrets.Required[i].Name = sub(cc.Spec.Secrets.Required[i].Name, vars)
	}
	for i := range cc.Spec.Secrets.Conditional {
		for j := range cc.Spec.Secrets.Conditional[i].Secrets {
			cc.Spec.Secrets.Conditional[i].Secrets[j].Name = sub(cc.Spec.Secrets.Conditional[i].Secrets[j].Name, vars)
		}
	}
}

func sub(s string, vars map[string]string) string {
	for k, v := range vars {
		s = strings.ReplaceAll(s, "${"+k+"}", v)
	}
	return s
}
