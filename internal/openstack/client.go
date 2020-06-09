package openstack

import (
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/gophercloud/gophercloud/pagination"
)

func GetComputeProvider(opts gophercloud.AuthOptions) (*gophercloud.ServiceClient, error) {
	provider, err := openstack.AuthenticatedClient(opts)
	if err != nil {
		return nil, err
	}

	computeProvider, err := openstack.NewComputeV2(provider, gophercloud.EndpointOpts{})
	if err != nil {
		return nil, err
	}

	return computeProvider, nil
}

func GetServers(computeProvider *gophercloud.ServiceClient) (s []servers.Server, err error) {
	listOpts := servers.ListOpts{}
	pager := servers.List(computeProvider, listOpts)
	err = pager.EachPage(func(page pagination.Page) (bool, error) {
		serverList, err := servers.ExtractServers(page)
		if err != nil {
			return false, err
		}
		s = append(s, serverList...)
		return true, nil
	})
	return s, err
}
