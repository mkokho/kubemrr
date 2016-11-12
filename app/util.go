package app

import (
	"fmt"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/rpc"
	"net/url"
	"os"
	"os/user"
)

func AddCommonFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("address", "a", "127.0.0.1", "The IP address where mirror is accessible")
	cmd.Flags().IntP("port", "p", 33033, "The port on which mirror is accessible")
	cmd.Flags().BoolP("verbose", "v", false, "Enables verbose output")
}

func RunCommon(cmd *cobra.Command) error {
	isVerbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return err
	} else if isVerbose {
		enableDebug()
	}
	return nil
}

func GetBind(cmd *cobra.Command) (string, error) {
	address, err := cmd.Flags().GetString("address")
	if err != nil {
		return "", err
	}

	port, err := cmd.Flags().GetInt("port")
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s:%d", address, port), nil
}

type Factory interface {
	KubeClient(baseUrl *url.URL) KubeClient
	MrrClient(bind string) (MrrClient, error)
	MrrCache() *MrrCache
	Serve(l net.Listener, c *MrrCache) error
	HomeKubeconfig() (Config, error)
	StdOut() io.Writer
}

type DefaultFactory struct {
	kubeconfig *Config
	stdOut     io.Writer
}

func NewFactory(stdOut io.Writer, kubeconfig *Config) Factory {
	return &DefaultFactory{
		stdOut:     stdOut,
		kubeconfig: kubeconfig,
	}
}

func (f *DefaultFactory) MrrClient(address string) (MrrClient, error) {
	return NewMrrClient(address)
}

func (f *DefaultFactory) StdOut() io.Writer {
	if f.stdOut == nil {
		return os.Stdout
	} else {
		return f.stdOut
	}
}

func (f *DefaultFactory) MrrCache() *MrrCache {
	return NewMrrCache()
}

func (f *DefaultFactory) KubeClient(url *url.URL) KubeClient {
	return NewKubeClient(url)
}

func (f *DefaultFactory) Serve(l net.Listener, cache *MrrCache) error {
	rpc.Register(cache)
	rpc.HandleHTTP()
	return http.Serve(l, nil)
}

func (f *DefaultFactory) HomeKubeconfig() (Config, error) {
	if f.kubeconfig != nil {
		return *f.kubeconfig, nil
	}

	usr, err := user.Current()
	if err != nil {
		return Config{}, err
	}
	return parseKubeConfig(usr.HomeDir + "/.kube/config")
}

func parseKubeConfig(filename string) (Config, error) {
	res := Config{}
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		return res, fmt.Errorf("could not read file %s: %s", filename, err)
	}

	err = yaml.Unmarshal(raw, &res)
	if err != nil {
		return res, fmt.Errorf("could not parse file %s: %s", filename, err)
	}

	return res, nil
}

type TestFactory struct {
	mrrClient   MrrClient
	mrrCache    *MrrCache
	kubeClients map[string]*TestKubeClient
	kubeconfig  Config
	stdOut      io.Writer
}

func NewTestFactory() *TestFactory {
	return &TestFactory{
		kubeClients: make(map[string]*TestKubeClient),
		mrrCache:    NewMrrCache(),
	}
}

func (f *TestFactory) MrrClient(address string) (MrrClient, error) {
	return f.mrrClient, nil
}

func (f *TestFactory) StdOut() io.Writer {
	if f.stdOut == nil {
		return os.Stdout
	} else {
		return f.stdOut
	}
}

func (f *TestFactory) MrrCache() *MrrCache {
	return f.mrrCache
}

func (f *TestFactory) Serve(l net.Listener, cache *MrrCache) error {
	return nil
}

func (f *TestFactory) HomeKubeconfig() (Config, error) {
	return f.kubeconfig, nil
}

func (f *TestFactory) KubeClient(url *url.URL) KubeClient {
	kc, ok := f.kubeClients[url.String()]
	if !ok {
		kc = NewTestKubeClient()
		kc.baseURL = url
		f.kubeClients[url.String()] = kc
	}
	return kc
}
