package app

import "strings"

type ObjectMeta struct {
	Name            string `json:"name,omitempty"`
	Namespace       string `json:"namespace,omitempty"`
	ResourceVersion string `json:"resourceVersion,omitempty"`
}

type TypeMeta struct {
	Kind string `json:"kind,omitempty"`
}

type KubeObject struct {
	TypeMeta   `json:",inline"`
	ObjectMeta `json:"metadata,omitempty"`
}

//KubeServer represents a Kubernetes API server which we ask for information
type KubeServer struct {
	URL string
}

//Config represent configuration written in .kube/config file
type Config struct {
	Clusters       []ClusterWrap `yaml:"clusters"`
	Contexts       []ContextWrap `yaml:"contexts"`
	CurrentContext string        `yaml:"current-context"`
}

type Cluster struct {
	Server string `yaml:"server"`
}

type ClusterWrap struct {
	Name    string  `yaml:"name"`
	Cluster Cluster `yaml:"cluster"`
}

type Context struct {
	Cluster   string `yaml:"cluster"`
	Namespace string `yaml:"namespace"`
}

type ContextWrap struct {
	Name    string  `json:"name"`
	Context Context `yaml:"context"`
}

func (c *Config) makeFilter() MrrFilter {
	var context Context
	for i := range c.Contexts {
		if c.Contexts[i].Name == c.CurrentContext {
			context = c.Contexts[i].Context
			break
		}
	}

	var cluster Cluster
	for i := range c.Clusters {
		if c.Clusters[i].Name == context.Cluster {
			cluster = c.Clusters[i].Cluster
			break
		}
	}

	return MrrFilter{
		Namespace: context.Namespace,
		Server:    cluster.urlWithoutPort(),
	}
}

func (c *Cluster) urlWithoutPort() string {
	i := strings.LastIndex(c.Server, ":")
	if i == -1 {
		return c.Server
	} else {
		return c.Server[:i]
	}
}
