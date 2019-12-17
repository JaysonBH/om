package commands_test

import (
	"fmt"
	"github.com/fatih/color"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/commands"
	"github.com/pivotal-cf/om/commands/fakes"
	"log"
)

var _ = Describe("ProductDiff", func() {

	var (
		logBuffer *gbytes.Buffer
		logger    *log.Logger
		service   *fakes.ProductDiffService
		err       error
	)

	BeforeEach(func() {
		service = &fakes.ProductDiffService{}
		logBuffer = gbytes.NewBuffer()
		logger = log.New(logBuffer, "", 0)
	})

	When("providing multiple products", func() {
		BeforeEach(func() {
			service.ProductDiffReturnsOnCall(0,
				api.ProductDiff{
					Manifest: api.ManifestDiff{
						Status: "different",
						Diff:   " properties:\n+  host: example.com\n-  host: localhost",
					},
					RuntimeConfigs: []api.RuntimeConfigsDiff{
						{
							Name:   "example-different-runtime-config",
							Status: "different",
							Diff:   " addons:\n - name: a-runtime-config\n   jobs:\n   - name: a-job\n     properties:\n+      timeout: 100\n-      timeout: 30",
						},
						{
							Name:   "example-same-runtime-config",
							Status: "same",
							Diff:   "",
						},
						{
							Name:   "example-also-different-runtime-config",
							Status: "different",
							Diff:   " addons:\n - name: another-runtime-config\n   jobs:\n   - name: another-job\n     properties:\n+      timeout: 110\n-      timeout: 31",
						},
					},
				}, nil)
			service.ProductDiffReturnsOnCall(1,
				api.ProductDiff{
					Manifest: api.ManifestDiff{
						Status: "same",
						Diff:   "",
					},
					RuntimeConfigs: []api.RuntimeConfigsDiff{
						{
							Name:   "example-different-runtime-config",
							Status: "same",
							Diff:   "",
						},
					},
				}, nil)
		})

		It("prints both product statuses", func() {
			diff := commands.NewProductDiff(service, logger)
			err = diff.Execute([]string{"--product", "example-product", "--product", "another-product"})
			Expect(err).NotTo(HaveOccurred())

			Expect(logBuffer).To(gbytes.Say("## Product Manifest for example-product"))
			Expect(logBuffer).To(gbytes.Say("properties:"))
			Expect(logBuffer).To(gbytes.Say("## Runtime Configs for example-product"))
			Expect(logBuffer).To(gbytes.Say("example-different-runtime-config"))

			Expect(logBuffer).To(gbytes.Say("## Product Manifest for another-product"))
			Expect(logBuffer).To(gbytes.Say("no changes"))
			Expect(logBuffer).To(gbytes.Say("## Runtime Configs for another-product"))
			Expect(logBuffer).To(gbytes.Say("no changes"))
		})
	})

	When("a product is provided", func() {
		When("there are both manifest and runtime config differences", func() {
			BeforeEach(func() {
				service.ProductDiffReturns(
					api.ProductDiff{
						Manifest: api.ManifestDiff{
							Status: "different",
							Diff:   " properties:\n+  host: example.com\n-  host: localhost",
						},
						RuntimeConfigs: []api.RuntimeConfigsDiff{
							{
								Name:   "example-different-runtime-config",
								Status: "different",
								Diff:   " addons:\n - name: a-runtime-config\n   jobs:\n   - name: a-job\n     properties:\n+      timeout: 100\n-      timeout: 30",
							},
							{
								Name:   "example-same-runtime-config",
								Status: "same",
								Diff:   "",
							},
							{
								Name:   "example-also-different-runtime-config",
								Status: "different",
								Diff:   " addons:\n - name: another-runtime-config\n   jobs:\n   - name: another-job\n     properties:\n+      timeout: 110\n-      timeout: 31",
							},
						},
					}, nil)
			})

			It("prints both", func() {
				//disable color for just this test;
				//we don't want to try to assemble this whole example with color
				color.NoColor = true
				defer func() { color.NoColor = false }()

				diff := commands.NewProductDiff(service, logger)
				err = diff.Execute([]string{"--product", "example-product"})
				Expect(err).NotTo(HaveOccurred())
				expectedOutput := `## Product Manifest for example-product

 properties:
+  host: example.com
-  host: localhost

## Runtime Configs for example-product

### example-different-runtime-config

 addons:
 - name: a-runtime-config
   jobs:
   - name: a-job
     properties:
+      timeout: 100
-      timeout: 30

### example-also-different-runtime-config

 addons:
 - name: another-runtime-config
   jobs:
   - name: another-job
     properties:
+      timeout: 110
-      timeout: 31
`
				Expect(string(logBuffer.Contents())).To(ContainSubstring(expectedOutput))
			})
			It("has colors on the diff", func() {
				diff := commands.NewProductDiff(service, logger)
				err = diff.Execute([]string{"--product", "example-product"})
				Expect(err).NotTo(HaveOccurred())

				bufferContents := string(logBuffer.Contents())

				Expect(bufferContents).To(ContainSubstring(color.GreenString("+  host: example.com")))
				Expect(bufferContents).To(ContainSubstring(color.RedString("-  host: localhost")))

				Expect(bufferContents).To(ContainSubstring(color.GreenString("+      timeout: 110")))
				Expect(bufferContents).To(ContainSubstring(color.RedString("-      timeout: 31")))
			})

		})

		When("there are product manifest changes only", func() {
			BeforeEach(func() {
				service.ProductDiffReturns(
					api.ProductDiff{
						Manifest: api.ManifestDiff{
							Status: "different",
							Diff:   " properties:\n+  host: example.com\n-  host: localhost",
						},
						RuntimeConfigs: []api.RuntimeConfigsDiff{
							{
								Name:   "example-different-runtime-config",
								Status: "same",
								Diff:   "",
							},
							{
								Name:   "example-same-runtime-config",
								Status: "same",
								Diff:   "",
							},
							{
								Name:   "example-also-different-runtime-config",
								Status: "same",
								Diff:   "",
							},
						},
					}, nil)
			})
			It("says there are no runtime config differences and prints manifest diffs", func() {
				diff := commands.NewProductDiff(service, logger)
				err = diff.Execute([]string{"--product", "example-product"})
				Expect(err).NotTo(HaveOccurred())
				Expect(logBuffer).To(gbytes.Say("## Product Manifest"))
				Expect(logBuffer).To(gbytes.Say("host: example.com"))
				Expect(logBuffer).To(gbytes.Say("## Runtime Configs"))
				Expect(logBuffer).To(gbytes.Say("no changes"))
			})
		})

		When("there are runtime config changes only", func() {
			BeforeEach(func() {
				service.ProductDiffReturns(
					api.ProductDiff{
						Manifest: api.ManifestDiff{
							Status: "same",
							Diff:   "",
						},
						RuntimeConfigs: []api.RuntimeConfigsDiff{
							{
								Name:   "example-different-runtime-config",
								Status: "different",
								Diff:   " addons:\n - name: a-runtime-config\n   jobs:\n   - name: a-job\n     properties:\n+      timeout: 100\n-      timeout: 30",
							},
						},
					}, nil)
			})

			It("says there are no product manifest differences and prints runtime config diffs", func() {
				diff := commands.NewProductDiff(service, logger)
				err = diff.Execute([]string{"--product", "example-product"})
				Expect(err).NotTo(HaveOccurred())
				Expect(logBuffer).To(gbytes.Say("## Product Manifest"))
				Expect(logBuffer).To(gbytes.Say("no changes"))
				Expect(logBuffer).To(gbytes.Say("## Runtime Configs"))
				Expect(logBuffer).To(gbytes.Say("timeout: 30"))

			})
		})

		When("there are neither manifest or runtime config changes", func() {
			BeforeEach(func() {
				service.ProductDiffReturns(
					api.ProductDiff{
						Manifest: api.ManifestDiff{
							Status: "same",
							Diff:   "",
						},
						RuntimeConfigs: []api.RuntimeConfigsDiff{
							{
								Name:   "example-different-runtime-config",
								Status: "same",
								Diff:   "",
							},
						},
					}, nil)
			})
			It("says there are no manifest differences and no runtime config diffs", func() {
				diff := commands.NewProductDiff(service, logger)
				err = diff.Execute([]string{"--product", "example-product"})
				Expect(err).NotTo(HaveOccurred())
				Expect(logBuffer).To(gbytes.Say("## Product Manifest"))
				Expect(logBuffer).To(gbytes.Say("no changes"))
				Expect(logBuffer).To(gbytes.Say("## Runtime Configs"))
				Expect(logBuffer).To(gbytes.Say("no changes"))
			})
		})

		When("the product is an addon tile with no manifest", func() {
			BeforeEach(func() {
				service.ProductDiffReturns(
					api.ProductDiff{
						Manifest: api.ManifestDiff{
							Status: "does_not_exist",
							Diff:   "",
						},
						RuntimeConfigs: []api.RuntimeConfigsDiff{
							{
								Name:   "example-different-runtime-config",
								Status: "different",
								Diff:   " addons:\n - name: a-runtime-config\n   jobs:\n   - name: a-job\n     properties:\n+      timeout: 100\n-      timeout: 90",
							},
						},
					}, nil)
			})

			It("says there is no manifest for the product and prints runtime config diffs", func() {
				diff := commands.NewProductDiff(service, logger)
				err = diff.Execute([]string{"--product", "example-product"})
				Expect(err).NotTo(HaveOccurred())
				Expect(logBuffer).To(gbytes.Say("## Product Manifest"))
				Expect(logBuffer).To(gbytes.Say("no manifest for this product"))
				Expect(logBuffer).To(gbytes.Say("## Runtime Configs"))
				Expect(logBuffer).To(gbytes.Say("timeout: 90"))
			})
		})

		When("the product is staged for initial installation", func() {
			BeforeEach(func() {
				service.ProductDiffReturns(
					api.ProductDiff{
						Manifest: api.ManifestDiff{
							Status: "to_be_installed",
							Diff:   "",
						},
						RuntimeConfigs: []api.RuntimeConfigsDiff{},
					}, nil)
			})

			It("says the product will be installed for the first time", func() {
				diff := commands.NewProductDiff(service, logger)
				err = diff.Execute([]string{"--product", "example-product"})
				Expect(err).NotTo(HaveOccurred())
				Expect(logBuffer).To(gbytes.Say("## Product Manifest"))
				Expect(logBuffer).To(gbytes.Say("This product is not yet deployed, so the product and runtime diffs are not available."))
				Expect(logBuffer).NotTo(gbytes.Say("## Runtime Configs"))
			})
		})

		PWhen("the product is staged for deletion", func() {
			// This case is actually a problem;
			// Products that are staged for deletion give a 404 when you hit their diff endpoint,
			// and don't show up in the list of staged products at all.
			// Needs discussion.
			BeforeEach(func() {
				service.ProductDiffReturns(
					api.ProductDiff{
						Manifest: api.ManifestDiff{
							Status: "to_be_deleted",
							Diff:   "",
						},
						RuntimeConfigs: []api.RuntimeConfigsDiff{},
					}, nil)
			})

			It("says the product will be deleted", func() {
				diff := commands.NewProductDiff(service, logger)
				err = diff.Execute([]string{"--product", "example-product"})
				Expect(err).NotTo(HaveOccurred())
				Expect(logBuffer).To(gbytes.Say("## Product Manifest"))
				Expect(logBuffer).To(gbytes.Say("This product will be deleted; product and runtime diffs are not available."))
				Expect(logBuffer).NotTo(gbytes.Say("## Runtime Configs"))
			})
		})
		When("there is an error from the diff service", func() {
			It("returns that error", func() {
				// setup
				service.ProductDiffReturns(
					api.ProductDiff{}, fmt.Errorf("too many cooks"))

				// execute
				diff := commands.NewProductDiff(service, logger)
				err = diff.Execute([]string{"--product", "err-product"})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("too many cooks"))
				Expect(service.ProductDiffArgsForCall(0)).To(Equal("err-product"))
			})
		})
	})

	When("no product is provided", func() {
		It("returns a validation error", func() {
			diff := commands.NewProductDiff(service, logger)
			err = diff.Execute([]string{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(`could not parse product-diff flags: missing required flag "--product"`))
		})
	})
})
