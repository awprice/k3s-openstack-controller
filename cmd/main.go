package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/awprice/k3s-openstack-controller/internal/k8s"
	"github.com/awprice/k3s-openstack-controller/internal/openstack"
	"github.com/gophercloud/gophercloud"
	"gopkg.in/alecthomas/kingpin.v2"
	"k8s.io/client-go/kubernetes"
)

var (
	app = kingpin.New("k3s-openstack-controller", "").DefaultEnvars()

	openstackIdentityEndpoint = app.Flag("openstack-identity-endpoint", "Openstack identity endpoint").Envar("OPENSTACK_IDENTITY_ENDPOINT").Required().String()
	openstackUsername         = app.Flag("openstack-username", "Openstack username").Envar("OPENSTACK_USERNAME").Required().String()
	openstackPassword         = app.Flag("openstack-password", "Openstack password").Envar("OPENSTACK_PASSWORD").Required().String()
	openstackDomainID         = app.Flag("openstack-domain-id", "Openstack domain ID").Envar("OPENSTACK_DOMAIN_ID").Required().String()
	openstackProjectName      = app.Flag("openstack-project-name", "Openstack project name").Envar("OPENSTACK_PROJECT_NAME").Required().String()
)

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	k8sConfig, err := k8s.GetConfig("")
	if err != nil {
		log.Fatal(err)
	}

	k8sClient, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		log.Fatal(err)
	}

	client := k8s.NewClient(k8sClient)

	opts := gophercloud.AuthOptions{
		IdentityEndpoint: *openstackIdentityEndpoint,
		Username:         *openstackUsername,
		Password:         *openstackPassword,
		DomainID:         *openstackDomainID,
		AllowReauth:      true,
		Scope: &gophercloud.AuthScope{
			ProjectName: *openstackProjectName,
			DomainID:    *openstackDomainID,
		},
	}

	computeProvider, err := openstack.GetComputeProvider(opts)
	if err != nil {
		log.Fatal(err)
	}

	for {
		if err := removeNodesNotInCloudProvider(client, computeProvider); err != nil {
			log.Printf("Error removing nodes not in cloud provider, %s", err.Error())
		}
		if err := removeOrphanedPods(client); err != nil {
			log.Printf("Error removing orphaned pods, %s", err.Error())
		}
		time.Sleep(10 * time.Second)
	}
}

func removeNodesNotInCloudProvider(client *k8s.Client, computeProvider *gophercloud.ServiceClient) error {
	nodes, err := client.GetNodes()
	if err != nil {
		return err
	}

	if len(nodes) == 0 {
		return fmt.Errorf("no nodes found")
	}

	servers, err := openstack.GetServers(computeProvider)
	if err != nil {
		return err
	}

	if len(servers) == 0 {
		return fmt.Errorf("no servers found")
	}

	for _, n := range nodes {
		uuid := strings.ToLower(n.Status.NodeInfo.SystemUUID)
		log.Printf("Checking if node %s, %s exists in cloud provider", n.Name, uuid)
		found := false
		for _, s := range servers {
			if strings.ToLower(s.ID) == uuid {
				found = true
				break
			}
		}
		if found {
			log.Println("Found in cloud provider, skipping")
			continue
		}

		log.Printf("Node %s does not exist, deleting...", n.Name)
		if err := client.DeleteNode(n.Name); err != nil {
			log.Printf("Error deleting node: %s", err.Error())
		}
	}

	return nil
}

func removeOrphanedPods(client *k8s.Client) error {
	pods, err := client.GetPods()
	if err != nil {
		return err
	}

	if len(pods) == 0 {
		return fmt.Errorf("no pods found")
	}

	nodes, err := client.GetNodes()
	if err != nil {
		return err
	}

	for _, p := range pods {
		// Ignore pods with no node
		if len(p.Spec.NodeName) == 0 {
			continue
		}
		log.Printf("Checking if node %s for pod %s/%s exists", p.Spec.NodeName, p.Namespace, p.Name)
		found := false
		for _, n := range nodes {
			if n.Name == p.Spec.NodeName {
				found = true
				break
			}
		}

		if found {
			log.Println("Found node, skipping")
			continue
		}

		log.Printf("Node %s does not exist, deleting pod %s/%s", p.Spec.NodeName, p.Namespace, p.Name)
		if err := client.DeletePod(p.Namespace, p.Name); err != nil {
			log.Printf("Error deleting pod: %s", err.Error())
		}
	}

	return nil
}
