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

package store

/*var images = []v1.Image{
	{
		ObjectMeta: metav1.ObjectMeta{
			Name: "a",
		},
		Spec: v1.ImageSpec{
			ID: "imagea",
		},
	},
	{
		ObjectMeta: metav1.ObjectMeta{
			Name: "b",
		},
		Spec: v1.ImageSpec{
			ID: "imageb",
		},
	},
	{
		ObjectMeta: metav1.ObjectMeta{
			Name: "c",
		},
		Spec: v1.ImageSpec{
			ID: "imagec",
		},
	},
	{
		ObjectMeta: metav1.ObjectMeta{
			Name: "d",
		},
		Spec: v1.ImageSpec{
			ID: "imaged",
		},
	},
}

var dirs = []string{
	imageDBRoot,
	common.DefaultTmpDir,
}

var platformMap = map[string]*v1.Platform{
	"amd64": {
		OS:           "linux",
		Architecture: "amd64",
	},
	"arm64": {
		OS:           "linux",
		Architecture: "arm64",
	},
	"arm": {
		OS:           "linux",
		Architecture: "arm",
		Variant:      "v6",
	},
}

func init() {
	for _, dir := range dirs {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			panic(err)
		}
	}
}

func TestImageStore_GetImage(t *testing.T) {
	skip.If(t, os.Getuid() != 0, "skipping test that requires root")

	is, err := NewDefaultImageStore()
	if err != nil {
		t.Error(err)
	}

	for _, image := range images {
		err = is.Save(image)
		if err != nil {
			t.Errorf("failed to save image %s, err: %s", image.Name, err)
		}
	}

	for _, image := range images {
		_, err = is.GetByID(image.Spec.ID)
		if err != nil {
			t.Errorf("failed to get image by id %s, err: %s", image.Spec.ID, err)
		}

		for _, plat := range platformMap {
			_, err = is.GetByName(image.Name, plat)
			if err != nil {
				t.Errorf("failed to get image by name %s, err: %s", image.Name, err)
			}
			_, err = is.GetImageMetadataItem(image.Name, plat)
			if err != nil {
				t.Errorf("failed to get image metadata item for %s, err: %s", image.Name, err)
			}
		}
	}
}

func TestImageStore_ImageMetadataItem(t *testing.T) {
	skip.If(t, os.Getuid() != 0, "skipping test that requires root")

	is, err := NewDefaultImageStore()
	if err != nil {
		t.Error(err)
	}

	for _, image := range images {
		for _, plat := range platformMap {
			err = is.SetImageMetadataItem(image.Name, &types.ManifestDescriptor{ID: image.Spec.ID, Platform: *plat})
			if err != nil {
				t.Errorf("failed to set image metadata for %s, err: %s", image.Name, err)
			}
		}
	}

	for _, image := range images {
		for _, plat := range platformMap {
			_, err = is.GetImageMetadataItem(image.Name, plat)
			if err != nil {
				t.Errorf("failed to set image metadata for %s, err: %s", image.Name, err)
			}
		}
	}
}
*/
