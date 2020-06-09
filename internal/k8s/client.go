package k8s

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func GetConfig(kubeconfig string) (*rest.Config, error) {
	if len(kubeconfig) > 0 {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	if len(os.Getenv("KUBECONFIG")) > 0 {
		return clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
	}
	if c, err := rest.InClusterConfig(); err == nil {
		return c, nil
	}
	if usr, err := user.Current(); err == nil {
		path := filepath.Join(usr.HomeDir, ".kube", "config")
		if c, err := clientcmd.BuildConfigFromFlags("", path); err == nil {
			return c, nil
		}
	}
	return nil, fmt.Errorf("could not locate a kubeconfig")
}

type Client struct {
	client kubernetes.Interface
}

func NewClient(client kubernetes.Interface) *Client {
	return &Client{client}
}

func (c *Client) GetNodes() ([]v1.Node, error) {
	nodeList, err := c.client.CoreV1().Nodes().List(metaV1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return nodeList.Items, nil
}

func (c *Client) DeleteNode(nodeName string) error {
	return c.client.CoreV1().Nodes().Delete(nodeName, &metaV1.DeleteOptions{})
}

func (c *Client) GetPods() ([]v1.Pod, error) {
	podList, err := c.client.CoreV1().Pods("").List(metaV1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return podList.Items, nil
}

func (c *Client) DeletePod(namespace, name string) error {
	return c.client.CoreV1().Pods(namespace).Delete(name, &metaV1.DeleteOptions{})
}
