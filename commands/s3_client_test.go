package commands_test

import (
	"archive/zip"
	"github.com/graymeta/stow"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"net/url"
	"strings"

	"io"
	"io/ioutil"
	"os"

	"github.com/pivotal-cf/om/commands"
	"github.com/pkg/errors"
)

var _ = Describe("S3Client", func() {
	Describe("GetAllProductVersions", func() {
		When("there are multiple files of the same 'version', differing by beta version", func() {
			var (
				stower *mockStower
				config commands.S3Configuration
			)

			BeforeEach(func() {
				itemsList := []mockItem{
					newMockItem("[product-slug,1.0.0-beta.1]someproductfile.zip"),
					newMockItem("[product-slug,1.0.0-beta.2]someproductfile.zip"),
					newMockItem("[product-slug,1.1.1]somefile-0.0.2.zip"),
					newMockItem("[product-slug,1.1.1]someotherfile-0.0.2.zip"),
				}

				stower = newMockStower(itemsList)
				config = commands.S3Configuration{
					Bucket:          "bucket",
					AccessKeyID:     "access-key-id",
					SecretAccessKey: "secret-access-key",
					RegionName:      "region",
					Endpoint:        "endpoint",
				}
			})

			It("reports all versions, including the beta versions", func() {
				client, err := commands.NewS3Client(stower, config, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())

				versions, err := client.GetAllProductVersions("product-slug")
				Expect(err).ToNot(HaveOccurred())

				Expect(versions).To(Equal([]string{
					"1.0.0-beta.1",
					"1.0.0-beta.2",
					"1.1.1",
				}))
			})
		})

		DescribeTable("the path variable", func(path string) {
			var (
				stower *mockStower
				config commands.S3Configuration
			)

			itemsList := []mockItem{
				newMockItem("/some-path/nested-path/[product-slug,8.8.8]someproductfile.zip"),
				newMockItem("/some-path/[product-slug,1.0.0-beta.1]someproductfile.zip"),
				newMockItem("some-path/[product-slug,1.2.3]someproductfile.zip"),
				newMockItem("[product-slug,7.7.7]someotherfile-0.0.2.zip"),
				newMockItem("/some-path/[product-slug,1.1.1]someotherfile-0.0.2.zip"),
				newMockItem("/some-path/[product-slug,1.1.1]with-another-right-bracket-]0.0.2.zip"),
			}

			stower = newMockStower(itemsList)
			config = commands.S3Configuration{
				Bucket:          "bucket",
				AccessKeyID:     "access-key-id",
				SecretAccessKey: "secret-access-key",
				RegionName:      "region",
				Endpoint:        "endpoint",
				Path:            path,
			}
			client, err := commands.NewS3Client(stower, config, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			versions, err := client.GetAllProductVersions("product-slug")
			Expect(err).ToNot(HaveOccurred())

			Expect(versions).To(Equal([]string{
				"1.0.0-beta.1",
				"1.2.3",
				"1.1.1",
			}))
		},
			Entry("with a leading and trailing slash", "/some-path/"),
			Entry("with a leading and without a trailing slash", "/some-path"),
			Entry("without a leading slash", "some-path/"),
			Entry("without a leading or trailing slash", "some-path"),
		)

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
					location: location,
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
				Expect(err.Error()).To(ContainSubstring("could not reach provided endpoint and bucket 'endpoint/bucket': expected element type <Error> but have StowErrorType"))
			})
		})

		When("zero files match the slug", func() {
			itemsList := []mockItem{
				newMockItem("product-slug-1.0.0-pcf-vsphere-2.1-build.341.ova"),
				newMockItem("product-slug-1.1.1-pcf-vsphere-2.1-build.348.ova"),
			}

			stower := newMockStower(itemsList)

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
					newMockItem("[product-slug,1.1.1]somefile-0.0.2.zip"),
				}
				stower := newMockStower(itemsList)
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

	Describe("GetLatestProductFile", func() {
		It("returns a file artifact", func() {
			itemsList := []mockItem{
				newMockItem("[product-slug,1.0.0]pcf-vsphere-2.1-build.341.ova"),
				newMockItem("[product-slug,1.1.1]pcf-vsphere-2.1-build.348.ova"),
			}

			stower := newMockStower(itemsList)
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
			Expect(fileArtifact.Name).To(Equal("[product-slug,1.1.1]pcf-vsphere-2.1-build.348.ova"))
		})

		It("errors when two files match the same glob", func() {
			itemsList := []mockItem{
				newMockItem("[product-slug,1.0.0]pcf-vsphere-2.1-build.341.ova"),
				newMockItem("[product-slug,1.1.1]pcf-vsphere-2.1-build.345.ova"),
				newMockItem("[product-slug,1.1.1]pcf-vsphere-2.1-build.348.ova"),
			}

			stower := newMockStower(itemsList)
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

		It("errors when zero prefixed files match the glob", func() {
			itemsList := []mockItem{
				newMockItem("[product-slug,1.0.0]pcf-vsphere-2.1-build.341.ova"),
				newMockItem("[product-slug,1.1.1]pcf-vsphere-2.1-build.345.ova"),
				newMockItem("[product-slug,1.1.1]pcf-vsphere-2.1-build.348.ova"),
			}

			stower := newMockStower(itemsList)
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

		DescribeTable("the item exists in the path in the bucket", func(path string) {
			itemsList := []mockItem{
				newMockItem("/some-path/nested/[product-slug,1.0.0]pcf-vsphere-2.1-build.341.ova"),
				newMockItem("/some-path/[product-slug,1.0.0]pcf-vsphere-2.1-build.341.ova"),
				newMockItem("some-path/[product-slug,1.1.1]pcf-vsphere-2.1-build.348.ova"),
				newMockItem("[product-slug,7.7.7]pcf-vsphere-2.1-build.348.ova"),
			}

			stower := newMockStower(itemsList)
			config := commands.S3Configuration{
				Bucket:          "bucket",
				AccessKeyID:     "access-key-id",
				SecretAccessKey: "secret-access-key",
				RegionName:      "region",
				Endpoint:        "endpoint",
				Path:            path,
			}

			client, err := commands.NewS3Client(stower, config, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())

			fileArtifact, err := client.GetLatestProductFile("product-slug", "1.1.1", "*vsphere*ova")
			Expect(err).ToNot(HaveOccurred())
			Expect(fileArtifact.Name).To(Equal("some-path/[product-slug,1.1.1]pcf-vsphere-2.1-build.348.ova"))
		},
			Entry("with a leading and trailing slash", "/some-path/"),
			Entry("with a leading and without a trailing slash", "/some-path"),
			Entry("without a leading slash", "some-path/"),
			Entry("without a leading or trailing slash", "some-path"),
		)
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

		It("writes to a file when the file exists", func() {
			item := newMockItem(file.Name())
			container := mockContainer{item: item}
			location := mockLocation{container: &container}
			stower := &mockStower{
				location:  location,
				itemsList: []mockItem{item},
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
				location: location,
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
			Expect(err.Error()).To(ContainSubstring("could not reach provided endpoint and bucket 'endpoint/bucket': expected element type <Error> but have StowErrorType"))
		})

		It("errors when cannot open file", func() {
			item := newMockItem(file.Name())
			item.fileError = errors.New("could not open file")
			container := mockContainer{item: item}
			location := mockLocation{container: &container}
			stower := &mockStower{
				location:  location,
				itemsList: []mockItem{item},
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

			err = client.DownloadProductToFile(&commands.FileArtifact{Name: "don't care"}, file)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("GetLatestStemcellForProduct", func() {
		When("the s3 bucket has stemcells that product can used", func() {
			It("returns the latest stemcell", func() {
				exampleTileFileName := createPivotalFile("[example-product,1.0-build.0]example*pivotal", "./fixtures/example-product-xenial-97.28.yml")

				stower := &mockStower{
					itemsList: []mockItem{
						newMockItem("[ubuntu-xenial,97.28]stemcell.tgz"),
						newMockItem("[ubuntu-xenial,97.54]stemcell.tgz"),
						newMockItem("[ubuntu-xenial,97.10]stemcell.tgz"),
						newMockItem("[ubuntu-xenial,97.101]stemcell.tgz"),
						newMockItem("[ubuntu-xenial,97.asdf]stemcell.tgz"),
					},
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

				stemcell, err := client.GetLatestStemcellForProduct(nil, exampleTileFileName)
				Expect(err).ToNot(HaveOccurred())

				Expect(stemcell.Version).To(Equal("97.101"))
				Expect(stemcell.Slug).To(Equal("ubuntu-xenial"))
			})
		})

		Context("failure cases", func() {
			It("errors with malformed stemcell version in the product", func() {
				exampleTileFileName := createPivotalFile("[example-product,1.0-build.0]example*pivotal", "./fixtures/example-product-xenial-bad-version.yml")

				stower := &mockStower{
					itemsList: []mockItem{},
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

				_, err = client.GetLatestStemcellForProduct(nil, exampleTileFileName)
				Expect(err).To(MatchError("versioning of stemcell dependency in unexpected format: \"major.minor\" or \"major\". the following version could not be parsed: whoops"))
			})

			It("errors when the product file does not have stemcell information", func() {
				config := commands.S3Configuration{
					Bucket:          "bucket",
					AccessKeyID:     "access-key-id",
					SecretAccessKey: "secret-access-key",
					RegionName:      "region",
					Endpoint:        "endpoint",
				}
				client, err := commands.NewS3Client(nil, config, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())

				_, err = client.GetLatestStemcellForProduct(nil, "./fixtures/example-product.yml")
				Expect(err).To(HaveOccurred())
			})

			It("errors when there are no available stemcell versions on s3", func() {
				exampleTileFileName := createPivotalFile("[example-product,1.0-build.0]example*pivotal", "./fixtures/example-product-xenial-97.28.yml")

				stower := &mockStower{
					itemsList: []mockItem{},
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

				_, err = client.GetLatestStemcellForProduct(nil, exampleTileFileName)
				Expect(err).To(MatchError("could not find stemcells on s3: bucket contains no files"))
			})

			It("errors when cannot get latest stemcell version", func() {
				exampleTileFileName := createPivotalFile("[example-product,1.0-build.0]example*pivotal", "./fixtures/example-product-xenial-97.28.yml")

				stower := &mockStower{
					itemsList: []mockItem{
						newMockItem("[ubuntu-xenial,96.28]stemcell.tgz"),
						newMockItem("[ubuntu-xenial,96.54]stemcell.tgz"),
						newMockItem("[ubuntu-xenial,96.10]stemcell.tgz"),
					},
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

				_, err = client.GetLatestStemcellForProduct(nil, exampleTileFileName)
				Expect(err).To(MatchError("no versions could be found equal to or greater than 97.28"))
			})
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
		itemsList := []mockItem{{}}
		stower := newMockStower(itemsList)
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
	dialCallCount  int
	dialError      error
	containerError error
	itemError      error
	config         commands.Config
}

func newMockStower(itemsList []mockItem) *mockStower {
	return &mockStower{
		itemsList: itemsList,
	}
}

func (s *mockStower) Dial(kind string, config commands.Config) (stow.Location, error) {
	s.config = config
	s.dialCallCount++
	if s.dialError != nil {
		return nil, s.dialError
	}

	return s.location, nil
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
	idString  string
	fileError error
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

	return ioutil.NopCloser(strings.NewReader("hello world")), nil
}

func (m mockItem) ID() string {
	return m.idString
}

func (m mockItem) Size() (int64, error) {
	return 0, nil
}

func createPivotalFile(productFileName, metadataFilename string) string {
	tempfile, err := ioutil.TempFile("", productFileName)
	Expect(err).NotTo(HaveOccurred())

	zipper := zip.NewWriter(tempfile)
	file, err := zipper.Create("metadata/props.yml")
	Expect(err).NotTo(HaveOccurred())

	contents, err := ioutil.ReadFile(metadataFilename)
	Expect(err).NotTo(HaveOccurred())

	_, err = file.Write(contents)
	Expect(err).NotTo(HaveOccurred())

	Expect(zipper.Close()).NotTo(HaveOccurred())
	return tempfile.Name()
}
