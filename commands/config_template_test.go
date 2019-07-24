package commands_test

import (
	"errors"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
	"io/ioutil"
	"path/filepath"
)

var _ = Describe("ConfigTemplate", func() {
	var (
		command *commands.ConfigTemplate
	)

	createOutputDirectory := func() string {
		tempDir, err := ioutil.TempDir("", "")
		Expect(err).ToNot(HaveOccurred())

		return tempDir
	}

	Describe("Execute", func() {
		BeforeEach(func() {
			command = commands.NewConfigTemplate(func(*commands.ConfigTemplate) commands.MetadataProvider {
				f := &fakes.MetadataProvider{}
				f.MetadataBytesReturns([]byte(`{name: example-product, product_version: "1.1.1"}`), nil)
				return f
			})
		})

		Describe("upserting an entry in the output directory with template files", func() {
			When("the output directory does not exist", func() {
				It("returns an error indicating the path does not exist", func() {
					args := []string{
						"--output-directory", "/not/real/directory",
						"--pivnet-api-token", "b",
						"--pivnet-product-slug", "c",
						"--product-version", "d",
					}

					err := command.Execute(args)
					Expect(err).To(MatchError("output-directory does not exist: /not/real/directory"))
				})

			})

			When("the output directory already exists without the product's directory", func() {
				var (
					tempDir string
					args    []string
				)

				BeforeEach(func() {
					tempDir = createOutputDirectory()

					args = []string{
						"--output-directory", tempDir,
						"--pivnet-api-token", "b",
						"--pivnet-product-slug", "c",
						"--product-version", "d",
					}
				})

				It("creates nested subdirectories named by product slug and version", func() {
					err := command.Execute(args)
					Expect(err).ToNot(HaveOccurred())

					productDir := filepath.Join(tempDir, "example-product")
					versionDir := filepath.Join(productDir, "1.1.1")

					Expect(productDir).To(BeADirectory())
					Expect(versionDir).To(BeADirectory())
				})

				It("creates the various generated sub directories within the product directory", func() {
					err := command.Execute(args)
					Expect(err).ToNot(HaveOccurred())

					featuresDir := filepath.Join(tempDir, "example-product", "1.1.1", "features")
					Expect(featuresDir).To(BeADirectory())

					networkDir := filepath.Join(tempDir, "example-product", "1.1.1", "network")
					Expect(networkDir).To(BeADirectory())

					optionalDir := filepath.Join(tempDir, "example-product", "1.1.1", "optional")
					Expect(optionalDir).To(BeADirectory())

					resourceDir := filepath.Join(tempDir, "example-product", "1.1.1", "resource")
					Expect(resourceDir).To(BeADirectory())
				})

				It("creates the correct files", func() {
					err := command.Execute(args)
					Expect(err).ToNot(HaveOccurred())

					productDir := filepath.Join(tempDir, "example-product", "1.1.1")

					Expect(filepath.Join(productDir, "errand-vars.yml")).To(BeAnExistingFile())
					Expect(filepath.Join(productDir, "product.yml")).To(BeAnExistingFile())
					Expect(filepath.Join(productDir, "default-vars.yml")).To(BeAnExistingFile())
					Expect(filepath.Join(productDir, "required-vars.yml")).To(BeAnExistingFile())
					Expect(filepath.Join(productDir, "resource-vars.yml")).To(BeAnExistingFile())
				})
			})
		})
	})

	Describe("Usage", func() {
		BeforeEach(func() {
			command = commands.NewConfigTemplate(func(*commands.ConfigTemplate) commands.MetadataProvider {
				f := &fakes.MetadataProvider{}
				f.MetadataBytesReturns([]byte(`{name: example-product, product_version: "1.1.1"}`), nil)
				return f
			})
		})

		It("returns usage information for the command", func() {
			Expect(command.Usage()).To(Equal(jhanda.Usage{
				Description:      "**EXPERIMENTAL** this command generates a product configuration template from a .pivotal file on Pivnet",
				ShortDescription: "**EXPERIMENTAL** generates a config template from a Pivnet product",
				Flags:            command.Options,
			}))
		})
	})

	Describe("flag handling", func() {
		When("an unknown flag is provided", func() {
			BeforeEach(func() {
				command = commands.NewConfigTemplate(func(*commands.ConfigTemplate) commands.MetadataProvider {
					f := &fakes.MetadataProvider{}
					f.MetadataBytesReturns([]byte(`{name: example-product, product_version: "1.1.1"}`), nil)
					return f
				})
			})
			It("returns an error", func() {
				err := command.Execute([]string{"--invalid"})
				Expect(err).To(MatchError("could not parse config-template flags: flag provided but not defined: -invalid"))
				err = command.Execute([]string{"--unreal"})
				Expect(err).To(MatchError("could not parse config-template flags: flag provided but not defined: -unreal"))
			})
		})

		When("the cli args arg not provided", func() {
			BeforeEach(func() {
				command = commands.NewConfigTemplate(func(*commands.ConfigTemplate) commands.MetadataProvider {
					f := &fakes.MetadataProvider{}
					f.MetadataBytesReturns([]byte(`{name: example-product, product_version: "1.1.1"}`), nil)
					return f
				})
			})
			DescribeTable("returns an error", func(required string) {
				args := []string{
					"--output-directory", "a",
					"--pivnet-api-token", "b",
					"--pivnet-product-slug", "c",
					"--product-version", "d",
				}
				for i, value := range args {
					if value == required {
						args = append(args[0:i], args[i+2:]...)
						break
					}
				}
				err := command.Execute(args)
				Expect(err).To(MatchError(fmt.Sprintf("could not parse config-template flags: missing required flag \"%s\"", required)))
			},
				Entry("with output-directory", "--output-directory"),
				Entry("with pivnet-api-token", "--pivnet-api-token"),
				Entry("with pivnet-product-slug", "--pivnet-product-slug"),
				Entry("with product-version", "--product-version"),
			)
		})

		Describe("metadata extraction and parsing failures", func() {
			When("the metadata cannot be extracted", func() {
				BeforeEach(func() {
					command = commands.NewConfigTemplate(func(*commands.ConfigTemplate) commands.MetadataProvider {
						f := &fakes.MetadataProvider{}
						f.MetadataBytesReturns(nil, errors.New("cannot get metadata"))
						return f
					})
				})

				It("returns an error", func() {
					tempDir := createOutputDirectory()

					args := []string{
						"--output-directory", tempDir,
						"--pivnet-api-token", "b",
						"--pivnet-product-slug", "example-product",
						"--product-version", "1.1.1",
					}

					err := command.Execute(args)
					Expect(err).To(MatchError("error getting metadata for example-product at version 1.1.1: cannot get metadata"))
				})
			})
			When("The returned metadata's version is an empty string", func() {
				BeforeEach(func() {
					command = commands.NewConfigTemplate(func(*commands.ConfigTemplate) commands.MetadataProvider {
						f := &fakes.MetadataProvider{}
						f.MetadataBytesReturns([]byte(`{name: example-product, product_version: ""}`), nil)
						return f
					})
				})
				It("errors", func() {
					tempDir := createOutputDirectory()

					args := []string{
						"--output-directory", tempDir,
						"--pivnet-api-token", "b",
						"--pivnet-product-slug", "example-product",
						"--product-version", "1.1.1",
					}

					err := command.Execute(args)
					Expect(err).To(HaveOccurred())
				})
			})
		})
	})
})
