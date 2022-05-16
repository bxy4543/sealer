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

package buildimage

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	osi "github.com/sealerio/sealer/utils/os"

	"github.com/sealerio/sealer/build/layerutils/charts"
	manifest "github.com/sealerio/sealer/build/layerutils/manifests"
	"github.com/sealerio/sealer/common"
	"github.com/sealerio/sealer/pkg/image/save"
	"github.com/sealerio/sealer/pkg/runtime"
	v1 "github.com/sealerio/sealer/types/api/v1"
	"golang.org/x/sync/errgroup"
)

var (
	copyToManifests = "manifests"
	copyToChart     = "charts"
	copyToImageList = "imageList"
	dispatch        map[string]func(srcPath string) ([]string, error)
)

func init() {
	dispatch = map[string]func(srcPath string) ([]string, error){
		copyToManifests: parseYamlImages,
		copyToChart:     parseChartImages,
		copyToImageList: parseRawImageList,
	}
}

type registry struct {
	platform v1.Platform
	puller   save.ImageSave
}

func (r registry) Process(srcPath, rootfs string) error {
	eg, _ := errgroup.WithContext(context.Background())

	var images []string
	for t, p := range dispatch {
		dispatchType := t
		parse := p
		eg.Go(func() error {
			ima, err := parse(srcPath)
			if err != nil {
				return fmt.Errorf("failed to parse images from %s error is : %v", dispatchType, err)
			}
			images = append(images, ima...)
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return err
	}

	return r.puller.SaveImages(images, filepath.Join(rootfs, common.RegistryDirName), r.platform)
}

func NewRegistryDiffer(platform v1.Platform) Differ {
	ctx := context.Background()
	return registry{
		platform: platform,
		puller:   save.NewImageSaver(ctx),
	}
}

func parseChartImages(srcPath string) ([]string, error) {
	chartsPath := filepath.Join(srcPath, copyToChart)
	if !osi.IsFileExist(chartsPath) {
		return nil, nil
	}

	var images []string
	imageSearcher, err := charts.NewCharts()
	if err != nil {
		return nil, err
	}

	err = filepath.Walk(chartsPath, func(path string, f fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !f.IsDir() {
			return nil
		}

		if osi.IsFileExist(filepath.Join(path, "Chart.yaml")) && osi.IsFileExist(filepath.Join(path, "values.yaml")) &&
			osi.IsFileExist(filepath.Join(path, "templates")) {
			ima, err := imageSearcher.ListImages(path)
			if err != nil {
				return err
			}
			images = append(images, ima...)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return FormatImages(images), nil
}

func parseYamlImages(srcPath string) ([]string, error) {
	manifestsPath := filepath.Join(srcPath, copyToManifests)
	if !osi.IsFileExist(manifestsPath) {
		return nil, nil
	}
	var images []string

	imageSearcher, err := manifest.NewManifests()
	if err != nil {
		return nil, err
	}

	err = filepath.Walk(manifestsPath, func(path string, f fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if f.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(f.Name()))
		if ext != ".yaml" && ext != ".yml" && ext != ".tmpl" {
			return nil
		}

		ima, err := imageSearcher.ListImages(path)

		if err != nil {
			return err
		}
		images = append(images, ima...)
		return nil
	})

	if err != nil {
		return nil, err
	}
	return FormatImages(images), nil
}

func parseRawImageList(srcPath string) ([]string, error) {
	imageListFilePath := filepath.Join(srcPath, copyToManifests, copyToImageList)
	if !osi.IsFileExist(imageListFilePath) {
		return nil, nil
	}

	images, err := osi.NewFileReader(imageListFilePath).ReadLines()
	if err != nil {
		return nil, fmt.Errorf("failed to read file content %s:%v", imageListFilePath, err)
	}
	return FormatImages(images), nil
}

type metadata struct {
}

func (m metadata) Process(srcPath, rootfs string) error {
	// check "KubeVersion" of Chart.yaml under charts dir,to overwrite the metadata.
	kv := getKubeVersion(srcPath)
	if kv == "" {
		return nil
	}

	md, err := m.loadMetadata(srcPath, rootfs)
	if err != nil {
		return err
	}

	if md.KubeVersion == kv {
		return nil
	}
	md.KubeVersion = kv
	mf := filepath.Join(rootfs, common.DefaultMetadataName)
	if err = marshalJSONToFile(mf, md); err != nil {
		return fmt.Errorf("failed to set image Metadata file, err: %v", err)
	}

	return nil
}

func (m metadata) loadMetadata(srcPath, rootfs string) (*runtime.Metadata, error) {
	// if Metadata file existed in srcPath, load and marshal to check the legality of it's content.
	// if not, use rootfs Metadata.
	smd, err := runtime.LoadMetadata(srcPath)
	if err != nil {
		return nil, err
	}
	if smd != nil {
		return smd, nil
	}

	md, err := runtime.LoadMetadata(rootfs)
	if err != nil {
		return nil, err
	}

	if md != nil {
		return md, nil
	}
	return nil, fmt.Errorf("failed to load rootfs Metadata")
}

func NewMetadataDiffer() Differ {
	return metadata{}
}
