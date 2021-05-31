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
	"path/filepath"

	"github.com/alibaba/sealer/common"
	"github.com/alibaba/sealer/test/suites/image"
	"github.com/alibaba/sealer/test/testhelper/settings"
	"github.com/alibaba/sealer/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("sealer image", func() {

	/*Context("pull image", func() {
		pullImageNames := []string{
			"registry.cn-qingdao.aliyuncs.com/sealer-io/kubernetes:v1.19.9",
			"registry.cn-qingdao.aliyuncs.com/kubernetes:v1.19.9",
			"sealer-io/kubernetes:v1.19.9",
			"kubernetes:v1.19.9",
		}
		It("pull true image", func() {
			for i := range pullImageNames {
				By(fmt.Sprintf("pull image %s", pullImageNames[i]), func() {
					image.DoImageOps("pull", pullImageNames[i])
					image.DoImageOps("rmi", pullImageNames[i])
				})
			}
		})

		faultPullImageNames := []string{
			"registry.cn-qingdao.aliyuncs.com/sealer-io:latest",
			"registry.cn-qingdao.aliyuncs.com:latest",
			"sealer-io:latest",
		}
		It("pull fault image", func() {
			for i := range faultPullImageNames {
				By(fmt.Sprintf("pull fault image %s", faultPullImageNames[i]), func() {
					sess, err := testhelper.Start(fmt.Sprintf("%s pull %s", settings.DefaultSealerBin, faultPullImageNames[i]))
					Expect(err).NotTo(HaveOccurred())
					Eventually(sess, settings.MaxWaiteTime).ShouldNot(Exit(0))
				})
			}
		})
	})*/

	Context("remove image", func() {
		/*It(fmt.Sprintf("remove image %s", settings.TestImageName), func() {
			beforeDirMd5, err := utils.DirMD5(filepath.Dir(common.DefaultImageRootDir))
			Expect(err).NotTo(HaveOccurred())
			image.DoImageOps("pull", settings.TestImageName)
			image.DoImageOps("rmi", settings.TestImageName)
			afterDirMd5, err := utils.DirMD5(filepath.Dir(common.DefaultImageRootDir))
			Expect(err).NotTo(HaveOccurred())
			Expect(afterDirMd5).To(Equal(beforeDirMd5))
		})*/

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

	/*Context("push image", func() {
		BeforeEach(func() {
			registry.Login()
			image.DoImageOps("pull", settings.TestImageName)
		})
		AfterEach(func() {
			registry.Logout()
			image.DoImageOps("rmi", settings.TestImageName)
		})
		pushImageNames := []string{
			"registry.cn-qingdao.aliyuncs.com/sealer-io/e2e_image_test:v0.01",
			"sealer-io/e2e_image_test:v0.01",
			"e2e_image_test:v0.01",
		}
		It("push image to repository", func() {
			for i := range pushImageNames {
				By(fmt.Sprintf("push image %s", pushImageNames[i]), func() {
					image.TagImages(settings.TestImageName, pushImageNames[i])
					image.DoImageOps("push", pushImageNames[i])
					image.DoImageOps("rmi", pushImageNames[i])
				})
			}
		})
	})*/
})
