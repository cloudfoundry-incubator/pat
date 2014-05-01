package workloads_test

import (
	//"crypto/md5"
	"io/ioutil"
	"os"
	"path"
	"strings"

	. "github.com/cloudfoundry-incubator/pat/workloads"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("GCF Workloads", func() {
	var (
		srcDir string
		dstDir string
	)
	BeforeEach(func() {
		srcDir = path.Join(os.TempDir(), "src")
		dstDir = path.Join(os.TempDir(), "dst")
	})
	AfterEach(func() {
		os.RemoveAll(srcDir)
		os.RemoveAll(dstDir)
	})

	Describe("Generating and Pushing an app", func() {
		Context("CopyAndReplaceText", func() {
			It("Copies the directory structure", func() {
				os.MkdirAll(path.Join(srcDir, "subdir"), 0777)
				CopyAndReplaceText(srcDir, dstDir, "", "")

				info, err := os.Lstat(dstDir)
				subInfo, err2 := os.Lstat(path.Join(dstDir, "subdir"))

				Ω(err).ShouldNot(HaveOccurred())
				Ω(err2).ShouldNot(HaveOccurred())
				Ω(info.IsDir()).Should(Equal(true))
				Ω(subInfo.IsDir()).Should(Equal(true))
			})

			It("Copies any files contained the source directory or subdirectories", func() {
				os.MkdirAll(path.Join(srcDir, "subdir"), 0777)
				file, _ := os.Create(path.Join(srcDir, "test.txt"))
				file.WriteString("abc123")
				file.Close()
				subFile, _ := os.Create(path.Join(srcDir, "subdir", "subfile.txt"))
				subFile.WriteString("foobar")
				subFile.Close()

				CopyAndReplaceText(srcDir, dstDir, "", "")
				dstFile, err := ioutil.ReadFile(path.Join(dstDir, "test.txt"))
				dstSubfile, err2 := ioutil.ReadFile(path.Join(dstDir, "subdir", "subfile.txt"))

				Ω(err).ShouldNot(HaveOccurred())
				Ω(err2).ShouldNot(HaveOccurred())
				Ω(string(dstFile)).Should(Equal("abc123"))
				Ω(string(dstSubfile)).Should(Equal("foobar"))
			})

			It("Replaces the target text in any copied files", func() {
				os.MkdirAll(path.Join(srcDir, "subdir"), 0777)
				file, _ := os.Create(path.Join(srcDir, "test.txt"))
				file.WriteString("abc123")
				file.WriteString("$RANDOM_TEXT")
				file.Close()
				subFile, _ := os.Create(path.Join(srcDir, "subdir", "subfile.txt"))
				subFile.WriteString("foobar")
				subFile.Close()

				CopyAndReplaceText(srcDir, dstDir, "$RANDOM_TEXT", "qwerty")

				dstFile, err := ioutil.ReadFile(path.Join(dstDir, "test.txt"))
				dstSubfile, err2 := ioutil.ReadFile(path.Join(dstDir, "subdir", "subfile.txt"))

				Ω(err).ShouldNot(HaveOccurred())
				Ω(err2).ShouldNot(HaveOccurred())
				Ω(strings.Contains(string(dstFile), "qwerty")).Should(Equal(true))
				Ω(strings.Contains(string(dstSubfile), "qwerty")).Should(Equal(false))
			})
		})
	})
})
