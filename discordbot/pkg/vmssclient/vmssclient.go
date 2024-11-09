package vmssclient

import (
	"context"
	"encoding/base64"
	"fmt"
	"godin/pkg/utils"
	"log"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/melbahja/goph"
)

type VmssClientInterface interface {
	ScaleUp() error
	ScaleDown() error
}

type VmssClient struct {
	Client            *armcompute.VirtualMachineScaleSetsClient
	VmssName          string
	ResourceGroupName string
	Ip                string
}

func NewVmssClient(resourcegroupname, vmssname, subscriptionid, ip string) (VmssClientInterface, error) {
	cred, err := azidentity.NewManagedIdentityCredential(nil)
	if err != nil {
		log.Printf("error creating azure cred: %v", err)
		return nil, fmt.Errorf("error creating azure cred: %v", err)
	}
	client, err := armcompute.NewVirtualMachineScaleSetsClient(subscriptionid, cred, nil)
	if err != nil {
		log.Printf("error creating vmss client: %v", err)
		return nil, fmt.Errorf("error creating vmss client: %v", err)
	}
	return &VmssClient{
		Client:            client,
		VmssName:          vmssname,
		ResourceGroupName: resourcegroupname,
		Ip:                ip,
	}, nil
}

func (vc *VmssClient) getCapacity() (*int64, error) {
	vmss, err := vc.Client.Get(context.TODO(), vc.ResourceGroupName, vc.VmssName, nil)
	if err != nil {
		return nil, err
	}
	return vmss.SKU.Capacity, nil
}

func (vc *VmssClient) ScaleUp() error {
	currentCapacity, err := vc.getCapacity()
	if err != nil {
		return err
	}
	if *currentCapacity == 1 {
		return nil
	}

	params := armcompute.VirtualMachineScaleSetUpdate{
		SKU: &armcompute.SKU{
			Capacity: utils.ToPtr(int64(1)),
		},
	}
	poller, err := vc.Client.BeginUpdate(context.TODO(), vc.ResourceGroupName, vc.VmssName, params, nil)
	if err != nil {
		return err
	}
	pudOpts := runtime.PollUntilDoneOptions{
		Frequency: 5 * time.Second,
	}
	_, err = poller.PollUntilDone(context.TODO(), &pudOpts)
	if err != nil {
		return err
	}
	return nil
}

func (vc *VmssClient) execInVm(command string) error {
	privKey, err := base64.StdEncoding.DecodeString(os.Getenv("BASE64_SERVER_KEY"))
	if err != nil {
		return err
	}
	auth, err := goph.RawKey(string(privKey), "")
	if err != nil {
		return err
	}
	client, err := goph.NewUnknown("azureuser", vc.Ip, auth)
	if err != nil {
		return err
	}
	defer client.Close()

	_, err = client.Run(command)
	if err != nil {
		return err
	}
	return nil
}

func (vc *VmssClient) ScaleDown() error {
	currentCapacity, err := vc.getCapacity()
	if err != nil {
		return err
	}
	if *currentCapacity == 0 {
		return nil
	}

	err = vc.execInVm("sudo docker stop valheim-server")
	if err != nil {
		return fmt.Errorf("failed to stop valheim container: %v", err)
	}
	params := armcompute.VirtualMachineScaleSetUpdate{
		SKU: &armcompute.SKU{
			Capacity: utils.ToPtr(int64(0)),
		},
	}
	poller, err := vc.Client.BeginUpdate(context.TODO(), vc.ResourceGroupName, vc.VmssName, params, nil)
	if err != nil {
		return err
	}
	pudOpts := runtime.PollUntilDoneOptions{
		Frequency: 5 * time.Second,
	}
	_, err = poller.PollUntilDone(context.TODO(), &pudOpts)
	if err != nil {
		return err
	}
	return nil
}
