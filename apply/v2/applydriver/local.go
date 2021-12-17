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

package applydriver

import (
	"fmt"

	v2 "github.com/alibaba/sealer/types/api/v2"

	"github.com/alibaba/sealer/apply/v2/processor"
	"github.com/alibaba/sealer/common"
	"github.com/alibaba/sealer/logger"
	"github.com/alibaba/sealer/pkg/runtime"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/alibaba/sealer/client/k8s"
	"github.com/alibaba/sealer/image"
	"github.com/alibaba/sealer/pkg/filesystem"
	"github.com/alibaba/sealer/utils"
)

// Applier cloud builder using cloud provider to build a cluster image
type Applier struct {
	ClusterDesired *v2.Cluster
	ClusterCurrent *v2.Cluster
	ImageManager   image.Service
	FileSystem     filesystem.Interface
	Client         *k8s.Client
}

func (c *Applier) Delete() (err error) {
	t := metav1.Now()
	c.ClusterDesired.DeletionTimestamp = &t
	return c.deleteCluster()
}

// Apply different actions between ClusterDesired and ClusterCurrent.
func (c *Applier) Apply() (err error) {
	// first time to init cluster
	if !utils.IsFileExist(common.DefaultKubeConfigFile()) {
		if err = c.initCluster(); err != nil {
			return err
		}
	} else {
		if err = c.reconcileCluster(); err != nil {
			return err
		}
	}

	return utils.SaveClusterInfoToFile(c.ClusterDesired, c.ClusterDesired.Name)
}

func (c *Applier) fillClusterCurrent() error {
	currentCluster, err := GetCurrentCluster(c.Client)
	if err != nil {
		return errors.Wrap(err, "get current cluster failed")
	}
	if currentCluster != nil {
		c.ClusterCurrent = c.ClusterDesired.DeepCopy()
		c.ClusterCurrent.Spec.Hosts = currentCluster.Spec.Hosts
	}
	return nil
}

func (c *Applier) mountClusterImage() error {
	imageName := c.ClusterDesired.Spec.Image
	err := c.ImageManager.PullIfNotExist(imageName)
	if err != nil {
		return err
	}
	err = c.FileSystem.MountImage(c.ClusterDesired)
	if err != nil {
		return err
	}
	return nil
}

func (c *Applier) unMountClusterImage() error {
	return c.FileSystem.UnMountImage(c.ClusterDesired)
}

func (c *Applier) reconcileCluster() error {
	client, err := k8s.Newk8sClient()
	if err != nil {
		return err
	}
	c.Client = client

	if err := c.fillClusterCurrent(); err != nil {
		return err
	}

	if err := c.mountClusterImage(); err != nil {
		return err
	}
	defer func() {
		if err := c.unMountClusterImage(); err != nil {
			logger.Warn("failed to umount image %s, %v", c.ClusterDesired.ClusterName, err)
		}
	}()
	mj, md := utils.GetDiffHosts(c.ClusterCurrent.GetMasterIPList(), c.ClusterDesired.GetMasterIPList())
	nj, nd := utils.GetDiffHosts(c.ClusterCurrent.GetNodeIPList(), c.ClusterDesired.GetNodeIPList())

	if err := c.scaleCluster(mj, md, nj, nd); err != nil {
		return err
	}

	if err := c.upgradeCluster(mj, nj); err != nil {
		return err
	}

	return nil
}

func (c *Applier) scaleCluster(mj, md, nj, nd []string) error {
	if len(mj) == 0 && len(md) == 0 && len(nj) == 0 && len(nd) == 0 {
		return nil
	}

	logger.Info("Start to scale this cluster")

	scaleProcessor, err := processor.NewScaleProcessor(c.FileSystem, mj, md, nj, nd)
	if err != nil {
		return err
	}
	var cluster *v2.Cluster
	if !scaleProcessor.(processor.ScaleProcessor).IsScaleUp {
		c, err := runtime.DecodeCRDFromFile(common.GetClusterWorkClusterfile(c.ClusterDesired.Name), common.Cluster)
		if err != nil {
			return err
		} else if c != nil {
			cluster = c.(*v2.Cluster)
		}
	} else {
		cluster = c.ClusterDesired
	}
	err = scaleProcessor.Execute(cluster)
	if err != nil {
		return err
	}

	logger.Info("Succeeded in scaling this cluster")

	return nil
}

func (c *Applier) upgradeCluster(mj, nj []string) error {
	// use k8sClient to fetch current cluster version.
	info, err := c.Client.GetClusterVersion()
	if err != nil {
		return err
	}
	// fetch form exec machine
	runtimeInterface, err := runtime.NewDefaultRuntime(c.ClusterDesired, c.ClusterDesired.GetAnnotationsByKey(common.ClusterfileName))
	if err != nil {
		return fmt.Errorf("failed to init runtime, %v", err)
	}
	clusterMetadata, err := runtimeInterface.GetClusterMetadata()
	if err != nil {
		return fmt.Errorf("failed to get cluster metadata: %v", err)
	}
	if info.GitVersion == clusterMetadata.Version {
		return nil
	}

	logger.Info("Start to upgrade this cluster from version(%s) to version(%s)", info.GitVersion, clusterMetadata.Version)
	//if desiredMetadata.Version==""{
	//	//install app
	//}

	upgradeProcessor, err := processor.NewUpgradeProcessor(c.FileSystem, runtimeInterface, mj, nj)
	if err != nil {
		return err
	}
	err = upgradeProcessor.Execute(c.ClusterDesired)
	if err != nil {
		return err
	}

	logger.Info("Succeeded in upgrading current cluster from version(%s) to version(%s)", info.GitVersion, clusterMetadata.Version)

	return nil
}

func (c *Applier) initCluster() error {
	logger.Info("Start to create a new cluster")
	createProcessor, err := processor.NewCreateProcessor()
	if err != nil {
		return err
	}

	if err := createProcessor.Execute(c.ClusterDesired); err != nil {
		return err
	}

	logger.Info("Succeeded in creating a new cluster, enjoy it!")

	return nil
}

func (c *Applier) deleteCluster() error {
	logger.Info("Start to delete current cluster")
	deleteProcessor, err := processor.NewDeleteProcessor()
	if err != nil {
		return err
	}

	if err := deleteProcessor.Execute(c.ClusterDesired); err != nil {
		return err
	}

	logger.Info("Succeeded in deleting current cluster")

	return nil
}
