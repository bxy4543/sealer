// Copyright © 2021 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package plugin

import (
	goContext "context"
	"fmt"
	"time"

	"golang.org/x/net/context"

	"github.com/sealerio/sealer/pkg/client/k8s"
	"github.com/sirupsen/logrus"
)

type ClusterChecker struct {
	client *k8s.Client
}

func NewClusterCheckerPlugin() Interface {
	return &ClusterChecker{}
}

func init() {
	Register(ClusterCheckPlugin, NewClusterCheckerPlugin())
}

func (c *ClusterChecker) Run(context Context, phase Phase) error {
	if phase != PhasePreGuest || context.Plugin.Spec.Type != ClusterCheckPlugin {
		logrus.Debug("check cluster is PreGuest!")
		return nil
	}
	if err := c.waitClusterReady(goContext.TODO()); err != nil {
		return err
	}
	return nil
}

func (c *ClusterChecker) waitClusterReady(ctx goContext.Context) error {
	var clusterStatusChan = make(chan string)
	ctx, cancel := context.WithTimeout(ctx, 15*time.Minute)
	defer cancel()
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	go func(t *time.Ticker) {
		for {
			clusterStatus := c.getClusterStatus()
			clusterStatusChan <- clusterStatus
			<-t.C
		}
	}(ticker)
	for {
		select {
		case status := <-clusterStatusChan:
			if status == ClusterNotReady {
				logrus.Info("wait for the cluster to ready ")
			} else if status == ClusterReady {
				logrus.Info("cluster is ready now")
				return nil
			}
		case <-ctx.Done():
			return fmt.Errorf("cluster is not ready, please check")
		}
	}
}

func (c *ClusterChecker) getClusterStatus() string {
	k8sClient, err := k8s.Newk8sClient()
	c.client = k8sClient
	if err != nil {
		return ClusterNotReady
	}

	kubeSystemPodStatus, err := c.client.ListKubeSystemPodsStatus()
	if !kubeSystemPodStatus || err != nil {
		return ClusterNotReady
	}

	return ClusterReady
}
