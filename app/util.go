package app

import (
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"net/url"
	"os"
)

func AddCommonFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("address", "a", "127.0.0.1", "The IP address where mirror is accessible")
	cmd.Flags().IntP("port", "p", 33033, "The port on which mirror is accessible")
}

func GetBind(cmd *cobra.Command) string {
	address, err := cmd.Flags().GetString("address")
	if err != nil {
		log.Fatal(err)
	}

	port, err := cmd.Flags().GetInt("port")
	if err != nil {
		log.Fatal(err)
	}

	return fmt.Sprintf("%s:%d", address, port)
}

type Factory interface {
	KubeClient(baseUrl *url.URL) KubeClient
	MrrClient(bind string) (MrrClient, error)
	MrrCache() *MrrCache
	Serve(l net.Listener, c *MrrCache) error
	StdOut() io.Writer
	StdErr() io.Writer
}

type DefaultFactory struct{}

func (f *DefaultFactory) MrrClient(address string) (MrrClient, error) {
	return NewMrrClient(address)
}

func (f *DefaultFactory) StdOut() io.Writer {
	return os.Stdout
}

func (f *DefaultFactory) StdErr() io.Writer {
	return os.Stderr
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

type TestFactory struct {
	mrrClient MrrClient
	mrrCache  MrrCache
	stdOut    io.Writer
	stdErr    io.Writer
}

func (f *TestFactory) MrrClient(address string) (MrrClient, error) {
	return f.mrrClient, nil
}

func (f *TestFactory) StdOut() io.Writer {
	return f.stdOut
}

func (f *TestFactory) StdErr() io.Writer {
	return f.stdErr
}

func (f *TestFactory) MrrCache() *MrrCache {
	return NewMrrCache()
}

func (f *TestFactory) Serve(l net.Listener, cache *MrrCache) error {
	return nil
}

func (f *TestFactory) KubeClient(url *url.URL) KubeClient {
	return nil
}
