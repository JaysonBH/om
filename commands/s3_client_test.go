package commands_test

import (
	"net/url"

	"github.com/graymeta/stow"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"io"
	"io/ioutil"
	"os"

	"github.com/pivotal-cf/om/commands"
	"github.com/pkg/errors"
)

var _ = Describe("S3Client", func() {
	var callCount int
	BeforeEach(func() {
		callCount = 0
	})

	Context("GetAllProductVersions", func() {
		It("returns versions matching the slug", func() {
			itemsList := []mockItem{
				newMockItem("product-slug-1.0.0-alpha.preview+123.github_somefile-0.0.1.zip"),
				newMockItem("product-slug-1.1.1_somefile-0.0.2.zip"),
				newMockItem("another-slug-1.2.3_somefile-0.0.3.zip"),
				newMockItem("another-slug-1.1.1_somefile-0.0.4.zip"),
			}
			stower := newMockStower(itemsList, &callCount)
			config := commands.S3Configuration{
				Bucket:          "bucket",
				AccessKeyID:     "access-key-id",
				SecretAccessKey: "secret-access-key",
				RegionName:      "region",
				Endpoint:        "endpoint",
			}

			client, err := commands.NewS3Client(stower, config, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			versions, err := client.GetAllProductVersions("product-slug")
			Expect(err).ToNot(HaveOccurred())

			Expect(versions).To(Equal([]string{
				"1.0.0-alpha.preview+123.github",
				"1.1.1",
			}))
		})

		It("does not include multiple copies of the same version", func() {
			itemsList := []mockItem{
				newMockItem("product-slug-1.0.0-alpha.preview+123.github_somefile-0.0.1.zip"),
				newMockItem("product-slug-1.1.1_somefile-0.0.2.zip"),
				newMockItem("product-slug-1.1.1_someotherfile-0.0.2.zip"),
				newMockItem("another-slug-1.2.3_somefile-0.0.3.zip"),
				newMockItem("another-slug-1.1.1_somefile-0.0.4.zip"),
			}

			stower := newMockStower(itemsList, &callCount)
			config := commands.S3Configuration{
				Bucket:          "bucket",
				AccessKeyID:     "access-key-id",
				SecretAccessKey: "secret-access-key",
				RegionName:      "region",
				Endpoint:        "endpoint",
			}

			client, err := commands.NewS3Client(stower, config, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			versions, err := client.GetAllProductVersions("product-slug")
			Expect(err).ToNot(HaveOccurred())

			Expect(versions).To(Equal([]string{
				"1.0.0-alpha.preview+123.github",
				"1.1.1",
			}))
		})

		When("the container returns 'expected element type <Error>", func() {
			var (
				stower *mockStower
				config commands.S3Configuration
			)

			BeforeEach(func() {
				location := mockLocation{
					containerError: errors.New("expected element type <Error> but have StowErrorType"),
				}
				stower = &mockStower{
					dialCallCount: &callCount,
					location:      location,
				}
				config = commands.S3Configuration{
					Bucket:          "bucket",
					AccessKeyID:     "access-key-id",
					SecretAccessKey: "secret-access-key",
					RegionName:      "region",
					Endpoint:        "endpoint",
				}
			})
			It("returns an error, containing endpoint information, saying S3 could not be reached", func() {

				client, err := commands.NewS3Client(stower, config, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())

				_, err = client.GetAllProductVersions("someslug")
				Expect(err.Error()).To(ContainSubstring("Could not reach provided endpoint: 'endpoint': expected element type <Error> but have StowErrorType"))
			})
		})

		When("zero files match the slug", func() {
			itemsList := []mockItem{
				newMockItem("product-slug-1.0.0-pcf-vsphere-2.1-build.341.ova"),
				newMockItem("product-slug-1.1.1-pcf-vsphere-2.1-build.348.ova"),
			}

			stower := newMockStower(itemsList, &callCount)

			config := commands.S3Configuration{
				Bucket:          "bucket",
				AccessKeyID:     "access-key-id",
				SecretAccessKey: "secret-access-key",
				RegionName:      "region",
			}

			It("gives an error message indicating the key and value that were not matched", func() {
				client, err := commands.NewS3Client(stower, config, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())

				_, err = client.GetAllProductVersions("someslug")
				Expect(err.Error()).To(ContainSubstring("no files matching pivnet-product-slug someslug found"))
			})
		})

		When("configuring s3", func() {
			It("can support v2 signing", func() {
				itemsList := []mockItem{
					newMockItem("product-slug-1.0.0-alpha.preview+123.github_somefile-0.0.1.zip"),
					newMockItem("product-slug-1.1.1_somefile-0.0.2.zip"),
					newMockItem("another-slug-1.2.3_somefile-0.0.3.zip"),
					newMockItem("another-slug-1.1.1_somefile-0.0.4.zip"),
				}
				stower := newMockStower(itemsList, &callCount)
				config := commands.S3Configuration{
					Bucket:          "bucket",
					AccessKeyID:     "access-key-id",
					SecretAccessKey: "secret-access-key",
					RegionName:      "region",
					Endpoint:        "endpoint",
					EnableV2Signing: true,
				}

				client, err := commands.NewS3Client(stower, config, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())

				_, err = client.GetAllProductVersions("product-slug")
				Expect(err).ToNot(HaveOccurred())

				Expect(stower.config).ToNot(BeNil())
				actualValue, _ := stower.config.Config("v2_signing")
				Expect(actualValue).To(Equal("true"))
			})
		})
	})

	Context("GetLatestProductFile", func() {
		It("returns a file artifact", func() {
			itemsList := []mockItem{
				newMockItem("product-slug-1.0.0-pcf-vsphere-2.1-build.341.ova"),
				newMockItem("product-slug-1.1.1-pcf-vsphere-2.1-build.348.ova"),
			}

			stower := newMockStower(itemsList, &callCount)
			config := commands.S3Configuration{
				Bucket:          "bucket",
				AccessKeyID:     "access-key-id",
				SecretAccessKey: "secret-access-key",
				RegionName:      "region",
				Endpoint:        "endpoint",
			}

			client, err := commands.NewS3Client(stower, config, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			fileArtifact, err := client.GetLatestProductFile("product-slug", "1.1.1", "*vsphere*ova")
			Expect(err).ToNot(HaveOccurred())
			Expect(fileArtifact.Name).To(Equal("product-slug-1.1.1-pcf-vsphere-2.1-build.348.ova"))
		})

		It("errors when two files match the same glob", func() {
			itemsList := []mockItem{
				newMockItem("product-slug-1.0.0-pcf-vsphere-2.1-build.341.ova"),
				newMockItem("product-slug-1.1.1-pcf-vsphere-2.1-build.345.ova"),
				newMockItem("product-slug-1.1.1-pcf-vsphere-2.1-build.348.ova"),
			}

			stower := newMockStower(itemsList, &callCount)
			config := commands.S3Configuration{
				Bucket:          "bucket",
				AccessKeyID:     "access-key-id",
				SecretAccessKey: "secret-access-key",
				RegionName:      "region",
				Endpoint:        "endpoint",
			}

			client, err := commands.NewS3Client(stower, config, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			_, err = client.GetLatestProductFile("product-slug", "1.1.1", "*vsphere*ova")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("the glob '*vsphere*ova' matches multiple files. Write your glob to match exactly one of the following"))
		})

		It("errors when zero files match the same glob", func() {
			itemsList := []mockItem{
				newMockItem("product-slug-1.0.0-pcf-vsphere-2.1-build.341.ova"),
				newMockItem("product-slug-1.1.1-pcf-vsphere-2.1-build.345.ova"),
				newMockItem("product-slug-1.1.1-pcf-vsphere-2.1-build.348.ova"),
			}

			stower := newMockStower(itemsList, &callCount)
			config := commands.S3Configuration{
				Bucket:          "bucket",
				AccessKeyID:     "access-key-id",
				SecretAccessKey: "secret-access-key",
				RegionName:      "region",
				Endpoint:        "endpoint",
			}

			client, err := commands.NewS3Client(stower, config, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			_, err = client.GetLatestProductFile("product-slug", "1.1.1", "*.zip")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("the glob '*.zip' matches no file"))
		})
	})

	Describe("DownloadProductToFile", func() {
		var file *os.File
		var fileContents = "hello world"

		BeforeEach(func() {
			var err error
			file, err = ioutil.TempFile("", "")
			Expect(err).NotTo(HaveOccurred())

			_, err = file.WriteString(fileContents)
			Expect(err).NotTo(HaveOccurred())

			err = file.Close()
		})

		AfterEach(func() {
			err := os.Remove(file.Name())
			Expect(err).ToNot(HaveOccurred())
		})

		Describe("DownloadFile", func() {
			It("receives the contents of a file if the file exists", func() {
				item := newMockItem(file.Name())
				item.fakeFileName = file.Name()
				container := mockContainer{item: item}
				location := mockLocation{container: &container}
				stower := &mockStower{
					location:      location,
					itemsList:     []mockItem{item},
					dialCallCount: &callCount,
				}

				config := commands.S3Configuration{
					Bucket:          "bucket",
					AccessKeyID:     "access-key-id",
					SecretAccessKey: "secret-access-key",
					RegionName:      "region",
					Endpoint:        "endpoint",
				}
				client, err := commands.NewS3Client(stower, config, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())

				readCloser, _, err := client.DownloadFile(file.Name())
				Expect(err).ToNot(HaveOccurred())

				b, err := ioutil.ReadAll(readCloser)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(b)).To(Equal(fileContents))
			})

			It("errors when cannot open file", func() {
				item := newMockItem(file.Name())
				item.fileError = errors.New("could not open file")
				container := mockContainer{item: item}
				location := mockLocation{container: &container}
				stower := &mockStower{
					location:      location,
					itemsList:     []mockItem{item},
					dialCallCount: &callCount,
				}

				config := commands.S3Configuration{
					Bucket:          "bucket",
					AccessKeyID:     "access-key-id",
					SecretAccessKey: "secret-access-key",
					RegionName:      "region",
					Endpoint:        "endpoint",
				}
				client, err := commands.NewS3Client(stower, config, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())

				_, _, err = client.DownloadFile(file.Name())
				Expect(err).To(HaveOccurred())
			})
		})

		It("writes to a file when the file exists", func() {
			item := newMockItem(file.Name())
			item.fakeFileName = file.Name()
			container := mockContainer{item: item}
			location := mockLocation{container: &container}
			stower := &mockStower{
				location:      location,
				itemsList:     []mockItem{item},
				dialCallCount: &callCount,
			}

			config := commands.S3Configuration{
				Bucket:          "bucket",
				AccessKeyID:     "access-key-id",
				SecretAccessKey: "secret-access-key",
				RegionName:      "region",
				Endpoint:        "endpoint",
			}
			client, err := commands.NewS3Client(stower, config, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			file, err := ioutil.TempFile("", "")
			Expect(err).ToNot(HaveOccurred())

			err = client.DownloadProductToFile(&commands.FileArtifact{Name: "don't care"}, file)
			Expect(err).ToNot(HaveOccurred())

			contents, err := ioutil.ReadFile(file.Name())
			Expect(err).ToNot(HaveOccurred())
			Expect(contents).To(Equal([]byte(fileContents)))
		})

		It("returns a helpful error if the InvalidSignature is returned by container", func() {
			location := mockLocation{
				containerError: errors.New("expected element type <Error> but have StowErrorType"),
			}
			stower := &mockStower{
				dialCallCount: &callCount,
				location:      location,
			}
			config := commands.S3Configuration{
				Bucket:          "bucket",
				AccessKeyID:     "access-key-id",
				SecretAccessKey: "secret-access-key",
				RegionName:      "region",
				Endpoint:        "endpoint",
			}

			file, err := ioutil.TempFile("", "")
			Expect(err).ToNot(HaveOccurred())

			client, err := commands.NewS3Client(stower, config, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			err = client.DownloadProductToFile(&commands.FileArtifact{Name: "don't care"}, file)
			Expect(err.Error()).To(ContainSubstring("Could not reach provided endpoint: 'endpoint': expected element type <Error> but have StowErrorType"))
		})
	})

	Describe("property validation and defaults", func() {
		DescribeTable("required property validation", func(param string) {
			stower := &mockStower{}
			config := commands.S3Configuration{}
			_, err := commands.NewS3Client(stower, config, GinkgoWriter)
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(ContainSubstring("Field validation for '%s' failed on the 'required' tag", param))
		},
			Entry("requires Bucket", "Bucket"),
			Entry("requires AccessKeyID", "AccessKeyID"),
			Entry("requires SecretAccessKey", "SecretAccessKey"),
			Entry("requires RegionName", "RegionName"),
		)

		It("defaults optional properties", func() {
			config := commands.S3Configuration{
				Bucket:          "bucket",
				AccessKeyID:     "access-key-id",
				SecretAccessKey: "secret-access-key",
				RegionName:      "region",
				Endpoint:        "endpoint",
			}
			stower := &mockStower{itemsList: []mockItem{}}
			client, err := commands.NewS3Client(stower, config, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			retrievedDisableSSLValue, retrievedValuePresence := client.Config.Config("disable_ssl")
			Expect(retrievedValuePresence).To(Equal(true))
			Expect(retrievedDisableSSLValue).To(Equal("false"))
		})

		When("both region and endpoint are given", func() {
			It("returns an error if they do not match", func() {
				config := commands.S3Configuration{
					Bucket:          "bucket",
					AccessKeyID:     "access-key-id",
					SecretAccessKey: "secret-access-key",
					RegionName:      "wrongRegion",
					Endpoint:        "endpoint",
				}
				stower := &mockStower{itemsList: []mockItem{}}
				_, err := commands.NewS3Client(stower, config, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	It("returns an error on stower failure", func() {
		dialError := errors.New("dial error")
		itemsList := []mockItem{}
		stower := newMockStower(itemsList, &callCount)
		stower.dialError = dialError

		config := commands.S3Configuration{
			Bucket:          "bucket",
			AccessKeyID:     "access-key-id",
			SecretAccessKey: "secret-access-key",
			RegionName:      "region",
			Endpoint:        "endpoint",
		}

		client, err := commands.NewS3Client(stower, config, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())

		_, err = client.GetAllProductVersions("product-slug")
		Expect(err).To(HaveOccurred())
		Expect(err).To(Equal(dialError))
	})
})

type mockStower struct {
	itemsList      []mockItem
	location       mockLocation
	dialCallCount  *int
	dialError      error
	containerError error
	itemError      error
	config         commands.Config
}

func newMockStower(itemsList []mockItem, callCount *int) *mockStower {
	return &mockStower{
		itemsList:     itemsList,
		dialCallCount: callCount,
	}
}

func (s *mockStower) Dial(kind string, config commands.Config) (stow.Location, error) {
	s.config = config
	*s.dialCallCount++
	if s.dialError != nil {
		return nil, s.dialError
	}

	return s.location, nil
}

func (s *mockStower) Container(id string) (stow.Container, error) {
	if s.containerError != nil {
		return nil, s.containerError
	}
	return mockContainer{}, nil
}

func (s *mockStower) Item(id string) (stow.Item, error) {
	if s.itemError != nil {
		return nil, s.itemError
	}
	return mockItem{}, nil
}

func (s *mockStower) Walk(container stow.Container, prefix string, pageSize int, fn stow.WalkFunc) error {
	for _, item := range s.itemsList {
		fn(item, nil)
	}

	return nil
}

type mockLocation struct {
	io.Closer
	container      *mockContainer
	containerError error
}

func (m mockLocation) CreateContainer(name string) (stow.Container, error) {
	return mockContainer{}, nil
}
func (m mockLocation) Containers(prefix string, cursor string, count int) ([]stow.Container, string, error) {
	return []stow.Container{mockContainer{}}, "", nil
}
func (m mockLocation) Container(id string) (stow.Container, error) {
	if m.containerError != nil {
		return nil, m.containerError
	}
	return m.container, nil
}
func (m mockLocation) RemoveContainer(id string) error {
	return nil
}
func (m mockLocation) ItemByURL(url *url.URL) (stow.Item, error) {
	return mockItem{}, nil
}

type mockContainer struct {
	item mockItem
}

func (m mockContainer) ID() string {
	return ""
}
func (m mockContainer) Name() string {
	return ""
}
func (m mockContainer) Item(id string) (stow.Item, error) {
	return m.item, nil
}
func (m mockContainer) Items(prefix, cursor string, count int) ([]stow.Item, string, error) {
	return []stow.Item{mockItem{}}, "", nil
}
func (m mockContainer) RemoveItem(id string) error {
	return nil
}
func (m mockContainer) Put(name string, r io.Reader, size int64, metadata map[string]interface{}) (stow.Item, error) {
	return mockItem{}, nil
}

type mockItem struct {
	stow.Item
	idString     string
	fakeFileName string
	fileError    error
}

func newMockItem(idString string) mockItem {
	return mockItem{
		idString: idString,
	}
}

func (m mockItem) Open() (io.ReadCloser, error) {
	if m.fileError != nil {
		return nil, m.fileError
	}

	if m.fakeFileName != "" {
		reader, err := os.Open(m.fakeFileName)
		Expect(err).ToNot(HaveOccurred())
		return ioutil.NopCloser(reader), nil
	}

	return nil, nil
}

func (m mockItem) ID() string {
	return m.idString
}

func (m mockItem) Size() (int64, error) {
	return 0, nil
}
