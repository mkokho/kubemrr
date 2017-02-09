package app

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

type KubeServers []KubeServer

func (s KubeServers) Len() int {
	return len(s)
}

func (s KubeServers) Less(i, j int) bool {
	return s[i].URL < s[j].URL
}

func (s KubeServers) Swap(i, j int) {
	x := s[j]
	s[j] = s[i]
	s[i] = x
}

//Config represent configuration written in .kube/config file
type Config struct {
	Clusters       []ClusterWrap `yaml:"clusters"`
	Contexts       []ContextWrap `yaml:"contexts"`
	Users          []UserWrap    `yaml:"users"`
	CurrentContext string        `yaml:"current-context"`
}

type Cluster struct {
	Server               string `yaml:"server"`
	CertificateAuthority string `yaml:"certificate-authority"`
}

type ClusterWrap struct {
	Name    string  `yaml:"name"`
	Cluster Cluster `yaml:"cluster"`
}

type Context struct {
	Cluster   string `yaml:"cluster"`
	Namespace string `yaml:"namespace"`
	User      string `yaml:"user"`
}

type ContextWrap struct {
	Name    string  `yaml:"name"`
	Context Context `yaml:"context"`
}

type User struct {
	ClientCertificate string `yaml:"client-certificate"`
	ClientKey         string `yaml:"client-key"`
}

type UserWrap struct {
	Name string `yaml:"name"`
	User User   `yaml:"user"`
}

func (c *Config) makeFilter() MrrFilter {
	var context Context
	for i := range c.Contexts {
		if c.Contexts[i].Name == c.CurrentContext {
			context = c.Contexts[i].Context
			break
		}
	}

	cluster := c.getCluster(context.Cluster)

	return MrrFilter{
		Namespace: context.Namespace,
		Server:    cluster.Server,
	}
}

func (c *Config) getCluster(name string) Cluster {
	var cluster Cluster
	for i := range c.Clusters {
		if c.Clusters[i].Name == name {
			cluster = c.Clusters[i].Cluster
			break
		}
	}
	return cluster
}
