package api_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf/om/api"
	"github.com/pivotal-cf/om/api/fakes"
	"net/http"
)

var _ = FDescribe("Diff Service", func() {
	var (
		server  *ghttp.Server
		stderr  *fakes.Logger
		service api.Api
	)

	BeforeEach(func() {
		server = ghttp.NewServer()

		stderr = &fakes.Logger{}
		service = api.New(api.ApiInput{
			Client: httpClient{server.URL()},
			Logger: stderr,
		})
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("ProductDiff", func() {
		When("an existing product is specified", func() {
			BeforeEach(func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
						ghttp.RespondWith(http.StatusOK, `[{
							"type": "some-product",
							"guid": "some-staged-guid"
						}]`),
					),
				)
			})

			It("returns the diff for the manifest", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/products/some-staged-guid/diff"),
						ghttp.RespondWith(http.StatusOK, `{
							"manifest": {
								"status": "different",
								"diff": " properties:\n+  test: new-value\n-  test: old-value"
							},
							"runtime_configs": []
						}`),
					),
				)

				diff, err := service.ProductDiff("some-product")
				Expect(err).NotTo(HaveOccurred())
				Expect(diff).To(Equal(api.ProductDiff{
					Manifest: api.ManifestDiff{
						Status: "different",
						Diff: " properties:\n+  test: new-value\n-  test: old-value",
					},
					RuntimeConfigs: []api.RuntimeConfigsDiff{},
				}))
			})

			PIt("returns the diff for the runtime config", func() {
				// Currently just a copy of the above test, pended/not implemented
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/products/some-staged-guid/diff"),
						ghttp.RespondWith(http.StatusOK, `{
							"manifest": {
								"status": "different",
								"diff": " properties:\n+  test: new-value\n-  test: old-value"
							},
							"runtime_configs": []
						}`),
					),
				)

				diff, err := service.ProductDiff("some-product")
				Expect(err).NotTo(HaveOccurred())
				Expect(diff).To(Equal(api.ProductDiff{
					Manifest: api.ManifestDiff{
						Status: "different",
						Diff: " properties:\n+  test: new-value\n-  test: old-value",
					},
					RuntimeConfigs: []api.RuntimeConfigsDiff{},
				}))
			})

			When("there is no diff returned for the product manifest", func() {
				It("succeeds and reports no diff", func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("GET", "/api/v0/products/some-staged-guid/diff"),
							ghttp.RespondWith(http.StatusOK, `{
								"manifest": {
									"status": "same",
									"diff": ""
								},
								"runtime_configs": []
							}`),
						),
					)

					diff, err := service.ProductDiff("some-product")
					Expect(err).NotTo(HaveOccurred())
					Expect(diff.Manifest.Diff).To(HaveLen(0))
				})
			})

			PWhen("there is no diff returned for the runtime config", func() {
				It("succeeds and reports no diff", func() {

				})
			})
		})

		When("the list products endpoint returns an error", func() {
			It("returns an error", func() {
				server.Close()

				_, err := service.ProductDiff("some-product")
				Expect(err).To(MatchError(ContainSubstring("could not make request to staged-products endpoint: could not send api request to GET /api/v0/staged/products")))
			})
		})

		When("the specified product cannot be found", func() {
			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
						ghttp.RespondWith(http.StatusOK, `[]`),
					),
				)

				_, err := service.ProductDiff("some-product")
				Expect(err).To(MatchError(`could not find product "some-product"`))
			})
		})

		When("the client has an error during the request when hitting the product diff endpoint", func() {
			It("returns an error", func() {
				// This will be called twice; http.Transport retries when connections close unexpectedly
				server.RouteToHandler("GET", "/api/v0/products/some-staged-guid/diff",
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						server.CloseClientConnections()
					}),
				)

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
						ghttp.RespondWith(http.StatusOK, `[{
							"type": "some-product",
							"guid": "some-staged-guid"
						}]`),
					),
				)

				_, err := service.ProductDiff("some-product")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not request product diff: could not send api request to GET /api/v0/products/some-staged-guid/diff"))
			})
		})

		When("the product diff endpoint returns a non-200 status code", func() {
			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
						ghttp.RespondWith(http.StatusOK, `[{
							"type": "some-product",
							"guid": "some-staged-guid"
						}]`),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/products/some-staged-guid/diff"),
						ghttp.RespondWith(http.StatusTeapot, ``),
					),
				)

				_, err := service.ProductDiff("some-product")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not retrieve product diff: request failed: unexpected response from /api/v0/products/some-staged-guid/diff:\nHTTP/1.1 418 I'm a teapot"))
			})
		})

		When("the product diff endpoint returns invalid json", func() {
			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/staged/products"),
						ghttp.RespondWith(http.StatusOK, `[{
							"type": "some-product",
							"guid": "some-staged-guid"
						}]`),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v0/products/some-staged-guid/diff"),
						ghttp.RespondWith(http.StatusOK, `actuallynotokayblaglegarg`),
					),
				)

				_, err := service.ProductDiff("some-product")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not unmarshal product diff response: %s", "actuallynotokayblaglegarg"))
			})
		})
	})
})
