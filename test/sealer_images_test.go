package test

import (
	"fmt"
	"path/filepath"

	"github.com/alibaba/sealer/common"
	"github.com/alibaba/sealer/test/suites/image"
	"github.com/alibaba/sealer/test/suites/registry"
	"github.com/alibaba/sealer/test/testhelper"
	"github.com/alibaba/sealer/test/testhelper/settings"
	"github.com/alibaba/sealer/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("sealer image", func() {

	Context("pull image", func() {
		pullImageNames := []string{
			"registry.cn-qingdao.aliyuncs.com/sealer-io/kubernetes:v1.19.9",
			"registry.cn-qingdao.aliyuncs.com/kubernetes:v1.19.9",
			"sealer-io/kubernetes:v1.19.9",
			"kubernetes:v1.19.9",
		}
		for _, imageName := range pullImageNames {
			It(fmt.Sprintf("pull image %s", imageName), func() {
				sess, err := testhelper.Start(fmt.Sprintf("sealer pull %s", imageName))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess, settings.MaxWaiteTime).Should(Exit(0))
				image.DoImageOps("rmi", imageName)
			})
		}

		faultPullImageNames := []string{
			"registry.cn-qingdao.aliyuncs.com/sealer-io:latest",
			"registry.cn-qingdao.aliyuncs.com:latest",
			"sealer-io:latest",
		}
		for _, imageName := range faultPullImageNames {
			It(fmt.Sprintf("pull fault image %s", imageName), func() {
				sess, err := testhelper.Start(fmt.Sprintf("sealer pull %s", imageName))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess, settings.MaxWaiteTime).ShouldNot(Exit(0))
			})
		}
	})

	Context("remove image", func() {
		It(fmt.Sprintf("remove image %s", settings.TestImageName), func() {
			beforeDirMd5, _ := utils.DirMD5(filepath.Dir(common.DefaultImageRootDir))
			image.DoImageOps("pull", settings.TestImageName)
			sess, err := testhelper.Start(fmt.Sprintf("sealer rmi %s", settings.TestImageName))
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess, settings.MaxWaiteTime).Should(Exit(0))
			afterDirMd5, _ := utils.DirMD5(filepath.Dir(common.DefaultImageRootDir))
			Expect(afterDirMd5).To(Equal(beforeDirMd5))
		})

		It("remove tag image", func() {
			tagImageName := "e2e_image_test:v0.01"
			image.DoImageOps("pull", settings.TestImageName)
			beforeDirMd5, _ := utils.DirMD5(filepath.Dir(common.DefaultImageRootDir))

			image.TagImages(settings.TestImageName, tagImageName)

			sess, err := testhelper.Start(fmt.Sprintf("sealer rmi %s", tagImageName))
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess, settings.MaxWaiteTime).Should(Exit(0))
			afterDirMd5, _ := utils.DirMD5(filepath.Dir(common.DefaultImageRootDir))
			Expect(afterDirMd5).To(Equal(beforeDirMd5))

			sess, err = testhelper.Start(fmt.Sprintf("sealer rmi %s", settings.TestImageName))
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess, settings.MaxWaiteTime).Should(Exit(0))
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
		var pushImageNames = []string{
			"registry.cn-qingdao.aliyuncs.com/sealer-io/e2e_image_test:v0.01",
			"sealer-io/e2e_image_test:v0.01",
			"e2e_image_test:v0.01",
		}
		for _, pushImageName := range pushImageNames {
			It(fmt.Sprintf("push image %s", pushImageName), func() {
				image.TagImages(settings.TestImageName, pushImageName)
				sess, err := testhelper.Start(fmt.Sprintf("sealer push %s", pushImageName))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess, settings.MaxWaiteTime).Should(Exit(0))
				image.DoImageOps("rmi", pushImageName)
			})
		}
	})
})
