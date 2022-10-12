# controller-utils

[![FairwindsOps](https://circleci.com/gh/FairwindsOps/controller-utils.svg?style=svg)](https://circleci.com/gh/FairwindsOps/controller-utils)
[![Apache 2.0 license](https://img.shields.io/badge/license-Apache2-brightgreen.svg)](https://opensource.org/licenses/Apache-2.0)

This is a library of Go functions to assist in building Kubernetes Controllers.

The `pkg/controller` package contains the main functionality. `pkg/log` contains helpers around logging. You can pass in a [logr](https://github.com/go-logr/logr) object to control the logs of this library.

## Basic Usage
```go
package main

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/fairwindsops/controller-utils/pkg/controller"
)

func main() {
	dynamic, restMapper, err := getKubeClient()
	if err != nil {
		panic(err)
	}
	client := controller.Client{
		Context:    context.TODO(),
		Dynamic:    dynamic,
		RESTMapper: restMapper,
	}
	workloads, err := client.GetAllTopControllersSummary("")
	if err != nil {
		panic(err)
	}
	for _, workload := range workloads {
		ctrl := workload.TopController
		fmt.Println()
		fmt.Printf("Workload: %s/%s/%s\n", ctrl.GetKind(), ctrl.GetNamespace(), ctrl.GetName())
		fmt.Printf("  num pods: %d\n", workload.PodCount)
		fmt.Printf("  running: %d\n", workload.RunningPodCount)
		fmt.Printf("  podSpec: %#v\n", workload.PodSpec)
	}
}

func getKubeClient() (dynamic.Interface, meta.RESTMapper, error) {
	var restMapper meta.RESTMapper
	var dynamicClient dynamic.Interface
	kubeConf, configError := ctrl.GetConfig()
	if configError != nil {
		return dynamicClient, restMapper, configError
	}

	api, err := kubernetes.NewForConfig(kubeConf)
	if err != nil {
		return dynamicClient, restMapper, err
	}

	dynamicClient, err = dynamic.NewForConfig(kubeConf)
	if err != nil {
		return dynamicClient, restMapper, err
	}

	resources, err := restmapper.GetAPIGroupResources(api.Discovery())
	if err != nil {
		return dynamicClient, restMapper, err
	}
	restMapper = restmapper.NewDiscoveryRESTMapper(resources)
	return dynamicClient, restMapper, nil
}
```

<!-- Begin boilerplate -->
## Join the Fairwinds Open Source Community

The goal of the Fairwinds Community is to exchange ideas, influence the open source roadmap,
and network with fellow Kubernetes users.
[Chat with us on Slack](https://join.slack.com/t/fairwindscommunity/shared_invite/zt-e3c6vj4l-3lIH6dvKqzWII5fSSFDi1g)
[join the user group](https://www.fairwinds.com/open-source-software-user-group) to get involved!

<a href="https://www.fairwinds.com/t-shirt-offer?utm_source=controller-utils&utm_medium=controller-utils&utm_campaign=controller-utils-tshirt">
  <img src="https://www.fairwinds.com/hubfs/Doc_Banners/Fairwinds_OSS_User_Group_740x125_v6.png" alt="Love Fairwinds Open Source? Share your business email and job title and we'll send you a free Fairwinds t-shirt!" />
</a>

## Other Projects from Fairwinds

Enjoying controller-utils? Check out some of our other projects:
* [Polaris](https://github.com/FairwindsOps/Polaris) - Audit, enforce, and build policies for Kubernetes resources, including over 20 built-in checks for best practices
* [Goldilocks](https://github.com/FairwindsOps/Goldilocks) - Right-size your Kubernetes Deployments by compare your memory and CPU settings against actual usage
* [Pluto](https://github.com/FairwindsOps/Pluto) - Detect Kubernetes resources that have been deprecated or removed in future versions
* [Nova](https://github.com/FairwindsOps/Nova) - Check to see if any of your Helm charts have updates available
* [rbac-manager](https://github.com/FairwindsOps/rbac-manager) - Simplify the management of RBAC in your Kubernetes clusters
