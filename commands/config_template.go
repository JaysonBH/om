package commands

import (
	"fmt"
	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/configtemplate/generator"
	"github.com/pivotal-cf/om/configtemplate/metadata"
	"os"
)

type ConfigTemplate struct {
	environFunc   func() []string
	buildProvider buildProvider
	Options       struct {
		OutputDirectory   string `long:"output-directory" required:"true"`
		PivnetApiToken    string `long:"pivnet-api-token" required:"true"`
		PivnetProductSlug string `long:"pivnet-product-slug" required:"true"`
		ProductVersion    string `long:"product-version" required:"true"`
	}
}

//go:generate counterfeiter -o ./fakes/metadata_provider.go --fake-name MetadataProvider . MetadataProvider
type MetadataProvider interface {
	MetadataBytes() ([]byte, error)
}

var DefaultProvider = func(host string) func(c *ConfigTemplate) MetadataProvider {
	return func(c *ConfigTemplate) MetadataProvider {
		options := c.Options
		return metadata.NewPivnetProvider(
			host,
			options.PivnetApiToken,
			options.PivnetProductSlug,
			options.ProductVersion,
			"*.pivotal",
		)
	}
}

type buildProvider func(*ConfigTemplate) MetadataProvider

func NewConfigTemplate(bp buildProvider) *ConfigTemplate {
	return &ConfigTemplate{
		environFunc:   os.Environ,
		buildProvider: bp,
	}
}

//Execute - generates config template and ops files
func (c *ConfigTemplate) Execute(args []string) error {
	_, err := jhanda.Parse(
		&c.Options,
		args)
	if err != nil {
		return fmt.Errorf("could not parse config-template flags: %s", err.Error())
	}

	_, err = os.Stat(c.Options.OutputDirectory)
	if os.IsNotExist(err) {
		return fmt.Errorf("output-directory does not exist: %s", c.Options.OutputDirectory)
	}

	metadataSource := c.newMetadataSource()
	metadataBytes, err := metadataSource.MetadataBytes()
	if err != nil {
		return fmt.Errorf("error getting metadata for %s at version %s: %s", c.Options.PivnetProductSlug, c.Options.ProductVersion, err)
	}

	return generator.NewExecutor(
		metadataBytes,
		c.Options.OutputDirectory,
		false,
		true,
	).Generate()
}

func (c *ConfigTemplate) newMetadataSource() (metadataSource MetadataProvider) {
	return c.buildProvider(c)
}

func (c *ConfigTemplate) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This command generates a product configuration template from a .pivotal file on Pivnet",
		ShortDescription: "generates a config template from a Pivnet product",
		Flags:            c.Options,
	}
}
