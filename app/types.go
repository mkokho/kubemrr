package app

type ObjectMeta struct {
	Name            string `json:"name,omitempty"`
	Namespace       string `json:"namespace,omitempty"`
	ResourceVersion string `json:"resourceVersion,omitempty"`
}

type TypeMeta struct {
	Kind string `json:"kind,omitempty"`
}

type Service struct {
	TypeMeta   `json:",inline"`
	ObjectMeta `json:"metadata,omitempty"`
}

type Pod struct {
	TypeMeta   `json:",inline"`
	ObjectMeta `json:"metadata,omitempty"`
}

type Deployment struct {
	TypeMeta   `json:",inline"`
	ObjectMeta `json:"metadata,omitempty"`
}

type KubeObject struct {
	TypeMeta   `json:",inline"`
	ObjectMeta `json:"metadata,omitempty"`
}

func (p *Pod) untype() KubeObject {
	return KubeObject{TypeMeta{"pod"}, p.ObjectMeta}
}

func (p *Service) untype() KubeObject {
	return KubeObject{TypeMeta{"service"}, p.ObjectMeta}
}

func (p *Deployment) untype() KubeObject {
	return KubeObject{TypeMeta{"deployment"}, p.ObjectMeta}
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
