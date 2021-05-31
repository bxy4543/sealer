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

package test

import (
	"fmt"
	"path/filepath"

	"github.com/alibaba/sealer/test/testhelper"

	"github.com/alibaba/sealer/common"
	"github.com/alibaba/sealer/test/suites/image"
	"github.com/alibaba/sealer/test/suites/registry"
	"github.com/alibaba/sealer/test/testhelper/settings"
	"github.com/alibaba/sealer/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("sealer image", func() {

	Context("pull image", func() {
		for i := 0; i < len(settings.PullImageNames); i++ {
			switch i {
			case 0:
				It(fmt.Sprintf("pull image %s", settings.PullImageNames[0]), func() {
					image.DoImageOps("pull", settings.PullImageNames[0])
					image.DoImageOps("rmi", settings.PullImageNames[0])
				})
			case 1:
				It(fmt.Sprintf("pull image %s", settings.PullImageNames[1]), func() {
					image.DoImageOps("pull", settings.PullImageNames[1])
					image.DoImageOps("rmi", settings.PullImageNames[1])
				})
			case 2:
				It(fmt.Sprintf("pull image %s", settings.PullImageNames[2]), func() {
					image.DoImageOps("pull", settings.PullImageNames[2])
					image.DoImageOps("rmi", settings.PullImageNames[2])
				})
			case 3:
				It(fmt.Sprintf("pull image %s", settings.PullImageNames[3]), func() {
					image.DoImageOps("pull", settings.PullImageNames[3])
					image.DoImageOps("rmi", settings.PullImageNames[3])
				})
			}
		}

		faultPullImageNames := []string{
			"registry.cn-qingdao.aliyuncs.com/sealer-io:latest",
			"registry.cn-qingdao.aliyuncs.com:latest",
			"sealer-io:latest",
		}
		for i := range faultPullImageNames {
			It(fmt.Sprintf("pull fault image %s", faultPullImageNames[i]), func() {
				sess, err := testhelper.Start(fmt.Sprintf("%s pull %s", settings.SealerBinPath, faultPullImageNames[i]))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess, settings.MaxWaiteTime).ShouldNot(Exit(0))
			})
		}
	})

	Context("remove image", func() {
		It(fmt.Sprintf("remove image %s", settings.TestImageName), func() {
			beforeDirMd5, err := utils.DirMD5(filepath.Dir(common.DefaultImageRootDir))
			Expect(err).NotTo(HaveOccurred())
			image.DoImageOps("pull", settings.TestImageName)
			image.DoImageOps("rmi", settings.TestImageName)
			afterDirMd5, err := utils.DirMD5(filepath.Dir(common.DefaultImageRootDir))
			Expect(err).NotTo(HaveOccurred())
			Expect(afterDirMd5).To(Equal(beforeDirMd5))
		})

		It("remove tag image", func() {
			tagImageName := "e2e_image_test:v0.01"
			image.DoImageOps("pull", settings.TestImageName)
			beforeDirMd5, err := utils.DirMD5(filepath.Dir(common.DefaultImageRootDir))
			Expect(err).NotTo(HaveOccurred())
			image.TagImages(settings.TestImageName, tagImageName)

			image.DoImageOps("rmi", tagImageName)
			afterDirMd5, err := utils.DirMD5(filepath.Dir(common.DefaultImageRootDir))
			Expect(err).NotTo(HaveOccurred())
			Expect(afterDirMd5).To(Equal(beforeDirMd5))

			image.DoImageOps("rmi", settings.TestImageName)
		})
	})

	Context("push image", func() {
		BeforeEach(func() {
			registry.Login()
			image.DoImageOps("pull", settings.TestImageName)
		})
		AfterEach(func() {
			registry.Logout()
			image.DoImageOps("rmi", settings.TestImageName)
		})

		for i := range settings.PushImageNames {
			It(fmt.Sprintf("push image %s", settings.PushImageNames[i]), func() {
				image.TagImages(settings.TestImageName, settings.PushImageNames[i])
				image.DoImageOps("push", settings.PushImageNames[i])
				image.DoImageOps("rmi", settings.PushImageNames[i])
			})
		}
	})
})
