package app

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"io/ioutil"
)

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
	Server                   string `yaml:"server"`
	SkipVerify               bool   `yaml:"insecure-skip-tls-verify"`
	CertificateAuthority     string `yaml:"certificate-authority"`
	CertificateAuthorityData string `yaml:"certificate-authority-data"`
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

func NewConfigFromURL(url string) (*Config, error) {
	config := Config{}
	cl := ClusterWrap{url, Cluster{Server: url}}
	ctx := ContextWrap{url, Context{Cluster: cl.Name}}
	config.Clusters = append(config.Clusters, cl)
	config.Contexts = append(config.Contexts, ctx)
	config.CurrentContext = url
	return &config, nil
}

func (c *Config) makeFilter() MrrFilter {
	context := c.getCurrentContext()
	cluster := c.getCluster(context.Cluster)

	return MrrFilter{
		Namespace: context.Namespace,
		Server:    cluster.Server,
	}
}

func (c *Config) getCurrentContext() Context {
	var context Context
	for i := range c.Contexts {
		if c.Contexts[i].Name == c.CurrentContext {
			context = c.Contexts[i].Context
			break
		}
	}
	return context
}

func (c *Config) getContext(name string) *Context {
	var context *Context
	for i := range c.Contexts {
		if c.Contexts[i].Name == name {
			context = &c.Contexts[i].Context
			break
		}
	}
	return context
}

func (c *Config) getCurrentCluster() Cluster {
	return c.getCluster(c.getCurrentContext().Cluster)
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

func (c *Config) getUser(name string) User {
	var user User
	for i := range c.Users {
		if c.Users[i].Name == name {
			user = c.Users[i].User
			break
		}
	}
	return user
}

func (cfg *Config) GenerateTLSConfig() (*tls.Config, error) {
	context := cfg.getCurrentContext()
	c := cfg.getCluster(context.Cluster)
	u := cfg.getUser(context.User)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: c.SkipVerify,
	}

	if len(c.CertificateAuthority) > 0 {
		caCertPool := x509.NewCertPool()
		caCert, err := ioutil.ReadFile(c.CertificateAuthority)
		if err != nil {
			return nil, fmt.Errorf("unable to use specified CA cert %s: %s", c.CertificateAuthority, err)
		}
		ok := caCertPool.AppendCertsFromPEM(caCert)
		if !ok {
			return nil, fmt.Errorf("unable to parse CA cert %s", c.CertificateAuthority)
		}
		tlsConfig.RootCAs = caCertPool
	}

	if len(c.CertificateAuthorityData) > 0 {
		caCert, err := base64.StdEncoding.DecodeString(c.CertificateAuthorityData)
		if err != nil {
			return nil, fmt.Errorf("unable to base 64 decode CA data")
		}

		caCertPool := x509.NewCertPool()
		ok := caCertPool.AppendCertsFromPEM(caCert)
		if !ok {
			return nil, fmt.Errorf("unable to parse CA data")
		}

		tlsConfig.RootCAs = caCertPool
	}

	if len(u.ClientCertificate) > 0 && len(u.ClientKey) == 0 {
		return nil, fmt.Errorf("client cert file %q specified without client key file", u.ClientCertificate)
	} else if len(u.ClientKey) > 0 && len(u.ClientCertificate) == 0 {
		return nil, fmt.Errorf("client key file %q specified without client cert file", u.ClientKey)
	} else if len(u.ClientCertificate) > 0 && len(u.ClientKey) > 0 {
		cert, err := tls.LoadX509KeyPair(u.ClientCertificate, u.ClientKey)
		if err != nil {
			return nil, fmt.Errorf("unable to use specified client cert (%s) & key (%s): %s", u.ClientCertificate, u.ClientKey, err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}
	tlsConfig.BuildNameToCertificate()

	return tlsConfig, nil
}
