package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"regexp"
	"sort"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/pivotal-cf/go-pivnet"
	pivnetlog "github.com/pivotal-cf/go-pivnet/logger"
	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/validator"
	"github.com/pivotal-cf/pivnet-cli/filter"
	"github.com/pivotal-cf/pivnet-cli/gp"
)

const DownloadProductOutputFilename = "download-file.json"

type outputList struct {
	ProductPath     string `json:"product_path,omitempty"`
	ProductSlug     string `json:"product_slug,omitempty"`
	StemcellPath    string `json:"stemcell_path,omitempty"`
	StemcellVersion string `json:"stemcell_version,omitempty"`
}

//go:generate counterfeiter -o ./fakes/pivnet_downloader_service.go --fake-name PivnetDownloader . PivnetDownloader
type PivnetDownloader interface {
	ReleasesForProductSlug(productSlug string) ([]pivnet.Release, error)
	ReleaseForVersion(productSlug string, releaseVersion string) (pivnet.Release, error)
	ProductFilesForRelease(productSlug string, releaseID int) ([]pivnet.ProductFile, error)
	DownloadProductFile(location *os.File, productSlug string, releaseID int, productFileID int, progressWriter io.Writer) error
	ReleaseDependencies(productSlug string, releaseID int) ([]pivnet.ReleaseDependency, error)
}

type ProductDownloader interface {
	GetAllProductVersions(slug string) ([]string, error)
	GetLatestProductFile(slug, version, glob string) (*FileArtifact, error)
	DownloadProductToFile(fa *FileArtifact, file *os.File) error
	DownloadProductStemcell(fa *FileArtifact) (*stemcell, error)
}

type PivnetFactory func(config pivnet.ClientConfig, logger pivnetlog.Logger) PivnetDownloader

func DefaultPivnetFactory(config pivnet.ClientConfig, logger pivnetlog.Logger) PivnetDownloader {
	return gp.NewClient(config, logger)
}

type DownloadProduct struct {
	environFunc    func() []string
	logger         pivnetlog.Logger
	progressWriter io.Writer
	pivnetFactory  PivnetFactory
	filter         *filter.Filter
	client         ProductDownloader
	Options        struct {
		ConfigFile          string   `long:"config"                short:"c" description:"path to yml file for configuration (keys must match the following command line flags)"`
		VarsFile            []string `long:"vars-file"             short:"l" description:"Load variables from a YAML file"`
		VarsEnv             []string `long:"vars-env"                        description:"Load variables from environment variables matching the provided prefix (e.g.: 'MY' to load MY_var=value)"`
		PivnetToken         string   `long:"pivnet-api-token"      short:"t" description:"API token to use when interacting with Pivnet. Can be retrieved from your profile page in Pivnet." required:"true"`
		PivnetFileGlob      string   `long:"pivnet-file-glob"      short:"f" description:"Glob to match files within Pivotal Network product to be downloaded." required:"true"`
		PivnetProductSlug   string   `long:"pivnet-product-slug"   short:"p" description:"Path to product" required:"true"`
		ProductVersion      string   `long:"product-version"       short:"v" description:"version of the product-slug to download files from. Incompatible with --product-version-regex flag."`
		ProductVersionRegex string   `long:"product-version-regex" short:"r" description:"Regex pattern matching versions of the product-slug to download files from. Highest-versioned match will be used. Incompatible with --product-version flag."`
		OutputDir           string   `long:"output-directory"      short:"o" description:"Directory path to which the file will be outputted. File Name will be preserved from Pivotal Network" required:"true"`
		Stemcell            bool     `long:"download-stemcell"               description:"No-op for backwards compatibility"`
		StemcellIaas        string   `long:"stemcell-iaas"                   description:"Download the latest available stemcell for the product for the specified iaas. for example 'vsphere' or 'vcloud' or 'openstack' or 'google' or 'azure' or 'aws'"`
	}
}

func NewDownloadProduct(environFunc func() []string, logger pivnetlog.Logger, progressWriter io.Writer, factory PivnetFactory) DownloadProduct {
	return DownloadProduct{
		environFunc:    environFunc,
		logger:         logger,
		progressWriter: progressWriter,
		pivnetFactory:  factory,
	}
}

func (c DownloadProduct) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This command attempts to download a single product file from Pivotal Network. The API token used must be associated with a user account that has already accepted the EULA for the specified product",
		ShortDescription: "downloads a specified product file from Pivotal Network",
		Flags:            c.Options,
	}
}

func (c DownloadProduct) Execute(args []string) error {
	err := loadConfigFile(args, &c.Options, c.environFunc)
	if err != nil {
		return fmt.Errorf("could not parse download-product flags: %s", err)
	}

	if c.Options.ProductVersionRegex != "" && c.Options.ProductVersion != "" {
		return fmt.Errorf("cannot use both --product-version and --product-version-regex; please choose one or the other")
	}

	c.client = NewPivnetClient(c.logger, c.progressWriter, c.pivnetFactory, c.Options.PivnetToken)

	productVersion := c.Options.ProductVersion
	if c.Options.ProductVersionRegex != "" {
		re, err := regexp.Compile(c.Options.ProductVersionRegex)
		if err != nil {
			return fmt.Errorf("could not compile regex: %s: %s", c.Options.ProductVersionRegex, err)
		}

		productVersions, err := c.client.GetAllProductVersions(c.Options.PivnetProductSlug)
		if err != nil {
			return err
		}

		var versions version.Collection
		for _, productVersion := range productVersions {
			if !re.MatchString(productVersion) {
				continue
			}

			v, err := version.NewVersion(productVersion)
			if err != nil {
				c.logger.Info(fmt.Sprintf("could not parse version: %s", productVersion))
				continue
			}
			versions = append(versions, v)
		}

		sort.Sort(versions)

		productVersion = versions[len(versions)-1].Original()
	}

	productFileName, productFileArtifact, err := c.downloadProductFile(c.Options.PivnetProductSlug, productVersion, c.Options.PivnetFileGlob)
	if err != nil {
		return fmt.Errorf("could not download product: %s", err)
	}

	if c.Options.StemcellIaas == "" {
		return c.writeOutputFile(productFileName, "", "")
	}

	c.logger.Info("Downloading stemcell")

	nameParts := strings.Split(productFileName, ".")
	if nameParts[len(nameParts)-1] != "pivotal" {
		c.logger.Info("the downloaded file is not a .pivotal file. Not determining and fetching required stemcell.")
		return nil
	}

	stemcell, err := c.client.DownloadProductStemcell(productFileArtifact)
	if err != nil {
		return fmt.Errorf("could not information about stemcell: %s", err)
	}

	stemcellFileName, _, err := c.downloadProductFile(stemcell.slug, stemcell.version, fmt.Sprintf("*%s*", c.Options.StemcellIaas))
	if err != nil {
		return fmt.Errorf("could not download stemcell: %s", err)
	}

	return c.writeOutputFile(productFileName, stemcellFileName, stemcell.version)
}

func (c DownloadProduct) writeOutputFile(productFileName string, stemcellFileName string, stemcellVersion string) error {
	c.logger.Info(fmt.Sprintf("Writing a list of downloaded artifact to %s", DownloadProductOutputFilename))
	outputList := outputList{
		ProductPath:     productFileName,
		StemcellPath:    stemcellFileName,
		ProductSlug:     c.Options.PivnetProductSlug,
		StemcellVersion: stemcellVersion,
	}

	outputFile, err := os.Create(path.Join(c.Options.OutputDir, DownloadProductOutputFilename))
	if err != nil {
		return fmt.Errorf("could not create %s: %s", DownloadProductOutputFilename, err)
	}
	defer outputFile.Close()

	return json.NewEncoder(outputFile).Encode(outputList)
}

func (c *DownloadProduct) downloadProductFile(slug, version, glob string) (string, *FileArtifact, error) {
	fileArtifact, err := c.client.GetLatestProductFile(slug, version, glob)
	if err != nil {
		return "", nil, err
	}

	productFilePath := path.Join(c.Options.OutputDir, path.Base(fileArtifact.Name))
	exist, err := checkFileExists(productFilePath, fileArtifact.sha256)
	if err != nil {
		return productFilePath, nil, err
	}

	if exist {
		c.logger.Info(fmt.Sprintf("%s already exists, skip downloading", productFilePath))
		return productFilePath, nil, nil
	}

	productFile, err := os.Create(productFilePath)
	if err != nil {
		return "", nil, fmt.Errorf("could not create file %s: %s", productFilePath, err)
	}
	defer productFile.Close()

	return productFilePath, fileArtifact, c.client.DownloadProductToFile(fileArtifact, productFile)
}

func checkFileExists(path, expectedSum string) (bool, error) {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		} else {
			return false, fmt.Errorf("failed to get file information: %s", err)
		}
	}

	validate := validator.NewSHA256Calculator()
	sum, err := validate.Checksum(path)
	if err != nil {
		return false, fmt.Errorf("failed to calculate the checksum: %s", err)
	}

	return sum == expectedSum, nil
}
