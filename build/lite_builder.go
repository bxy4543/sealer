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

package build

import (
	"fmt"
	"path/filepath"

	"github.com/alibaba/sealer/image"
	"github.com/alibaba/sealer/utils/mount"

	manifest "github.com/alibaba/sealer/build/lite/manifests"

	"github.com/alibaba/sealer/build/lite/charts"
	"github.com/alibaba/sealer/build/lite/docker"
	"github.com/alibaba/sealer/common"
	"github.com/alibaba/sealer/filesystem"
	"github.com/alibaba/sealer/logger"
	v1 "github.com/alibaba/sealer/types/api/v1"
	"github.com/alibaba/sealer/utils"
)

type LiteBuilder struct {
	local *LocalBuilder
}

func (l *LiteBuilder) Build(name string, context string, kubefileName string) error {
	err := l.local.initBuilder(name, context, kubefileName)
	if err != nil {
		return err
	}

	pipLine, err := l.GetBuildPipeLine()
	if err != nil {
		return err
	}

	for _, f := range pipLine {
		if err = f(); err != nil {
			return err
		}
	}
	return nil
}

// load cluster file from disk
func (l *LiteBuilder) InitClusterFile() error {
	clusterFile := common.TmpClusterfile
	if !utils.IsFileExist(clusterFile) {
		rawClusterFile := GetRawClusterFile(l.local.Image)
		if rawClusterFile == "" {
			return fmt.Errorf("failed to get cluster file from context or base image")
		}
		err := utils.WriteFile(common.RawClusterfile, []byte(rawClusterFile))
		if err != nil {
			return err
		}
		clusterFile = common.RawClusterfile
	}
	var cluster v1.Cluster
	err := utils.UnmarshalYamlFile(clusterFile, &cluster)
	if err != nil {
		return fmt.Errorf("failed to read %s:%v", clusterFile, err)
	}
	l.local.Cluster = &cluster

	logger.Info("read cluster file %s success !", clusterFile)
	return nil
}

func (l *LiteBuilder) GetBuildPipeLine() ([]func() error, error) {
	var buildPipeline []func() error
	if err := l.local.InitImageSpec(); err != nil {
		return nil, err
	}

	buildPipeline = append(buildPipeline,
		l.PreCheck,
		l.local.PullBaseImageNotExist,
		l.InitClusterFile,
		l.MountImage,
		l.local.ExecBuild,
		l.local.UpdateImageMetadata,
		l.ReMountImage,
		l.InitDockerAndRegistry,
		l.CacheImageToRegistry,
		l.AddUpperLayerToImage,
		l.Clear,
	)
	return buildPipeline, nil
}

func (l *LiteBuilder) PreCheck() error {
	d := docker.Docker{}
	images, _ := d.ImagesList()
	if len(images) > 0 {
		logger.Warn("The image already exists on the host. Note that the existing image cannot be cached in registry")
	}
	return nil
}

func (l *LiteBuilder) ReMountImage() error {
	err := l.UnMountImage()
	if err != nil {
		return err
	}
	l.local.Cluster.Spec.Image = l.local.Config.ImageName
	return l.MountImage()
}

func (l *LiteBuilder) UnMountImage() error {
	var (
		FileSystem filesystem.Interface
		err        error
	)
	FileSystem, err = filesystem.NewFilesystem()
	if err != nil {
		logger.Warn(err)
		return err
	}
	return FileSystem.UnMountImage(l.local.Cluster)
}

func (l *LiteBuilder) MountImage() error {
	FileSystem, err := filesystem.NewFilesystem()
	if err != nil {
		return err
	}
	if err := FileSystem.MountImage(l.local.Cluster); err != nil {
		return err
	}
	return nil
}

func (l *LiteBuilder) AddUpperLayerToImage() error {
	var (
		err   error
		Image *v1.Image
	)
	m := filepath.Join(common.DefaultClusterBaseDir(l.local.Cluster.Name), "mount")
	err = mount.NewMountDriver().Unmount(m)
	if err != nil {
		return err
	}
	upper := filepath.Join(m, "upper")
	imageLayer := v1.Layer{
		Type:  "BASE",
		Value: "registry cache",
	}
	err = l.local.calculateLayerDigestAndPlaceIt(&imageLayer, upper)
	if err != nil {
		return err
	}
	Image, err = image.GetImageByName(l.local.Config.ImageName)
	if err != nil {
		return err
	}
	Image.Spec.Layers = append(Image.Spec.Layers, imageLayer)
	l.local.Image = Image
	err = l.local.updateImageIDAndSaveImage()
	if err != nil {
		return err
	}
	return nil
}

func (l *LiteBuilder) Clear() error {
	return utils.CleanFiles(common.RawClusterfile, common.DefaultClusterBaseDir(l.local.Cluster.Name))
}

func (l *LiteBuilder) InitDockerAndRegistry() error {
	mount := filepath.Join(common.DefaultClusterBaseDir(l.local.Cluster.Name), "mount")
	cmd := "cd %s  && chmod +x scripts/* && cd scripts && sh docker.sh && sh init-registry.sh 5000 %s"
	r, err := utils.CmdOutput("sh", "-c", fmt.Sprintf(cmd, mount, filepath.Join(mount, "registry")))
	if err != nil {
		logger.Error(fmt.Sprintf("Init docker and registry failed: %v", err))
		return err
	}
	logger.Info(string(r))
	return nil
}

func (l *LiteBuilder) CacheImageToRegistry() error {
	var images []string
	var err error
	d := docker.Docker{}
	c := charts.Charts{}
	m := manifest.Manifests{}
	imageList := filepath.Join(common.DefaultClusterBaseDir(l.local.Cluster.Name), "mount", "manifests", "imageList")
	if utils.IsExist(imageList) {
		images, err = utils.ReadLines(imageList)
	}
	if helmImages, err := c.ListImages(l.local.Cluster.Name); err == nil {
		images = append(images, helmImages...)
	}
	if manifestImages, err := m.ListImages(l.local.Cluster.Name); err == nil {
		images = append(images, manifestImages...)
	}
	if err != nil {
		return err
	}
	d.ImagesPull(images)
	return nil
}

func NewLiteBuilder(config *Config) (Interface, error) {
	localBuilder, err := NewLocalBuilder(config)
	if err != nil {
		return nil, err
	}
	return &LiteBuilder{
		local: localBuilder.(*LocalBuilder),
	}, nil
}
