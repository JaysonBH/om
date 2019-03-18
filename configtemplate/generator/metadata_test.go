package generator_test

import (
	"github.com/onsi/ginkgo/extensions/table"
	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotalservices/tile-config-generator/generator"
)

var _ = Describe("Metadata", func() {
	Context("UsesServiceNetwork", func() {
		It("Should use service network", func() {
			fileData, err := ioutil.ReadFile("fixtures/p_healthwatch.yml")
			Expect(err).ShouldNot(HaveOccurred())
			metadata, err := generator.NewMetadata(fileData)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(metadata.UsesServiceNetwork()).Should(BeTrue())
		})

		It("Should not service network", func() {
			fileData, err := ioutil.ReadFile("fixtures/pas.yml")
			Expect(err).ShouldNot(HaveOccurred())
			metadata, err := generator.NewMetadata(fileData)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(metadata.UsesServiceNetwork()).Should(BeFalse())
		})

	})

	Context("GetPropertyMetadata", func() {
		It("returns a non-job configurable property", func() {
			fileData, err := ioutil.ReadFile("fixtures/p_healthwatch.yml")
			Expect(err).ShouldNot(HaveOccurred())
			metadata, err := generator.NewMetadata(fileData)
			Expect(err).ShouldNot(HaveOccurred())
			property, err := metadata.GetPropertyMetadata(".properties.opsman")
			Expect(err).ShouldNot(HaveOccurred())
			Expect(property.Name).Should(Equal("opsman"))
		})

		It("returns a job configurable property", func() {
			fileData, err := ioutil.ReadFile("fixtures/p_healthwatch.yml")
			Expect(err).ShouldNot(HaveOccurred())
			metadata, err := generator.NewMetadata(fileData)
			Expect(err).ShouldNot(HaveOccurred())
			property, err := metadata.GetPropertyMetadata(".healthwatch-forwarder.foundation_name")
			Expect(err).ShouldNot(HaveOccurred())
			//Expect(property).ShouldNot(BeNil())
			Expect(property.Name).Should(Equal("foundation_name"))
		})
	})

	Context("Product Name", func() {
		It("Should return cf as product name", func() {
			fileData, err := ioutil.ReadFile("fixtures/pas.yml")
			Expect(err).ShouldNot(HaveOccurred())
			metadata, err := generator.NewMetadata(fileData)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(metadata.ProductName()).Should(BeEquivalentTo("cf"))
		})

		It("Should return cf as product name", func() {
			fileData, err := ioutil.ReadFile("fixtures/srt.yml")
			Expect(err).ShouldNot(HaveOccurred())
			metadata, err := generator.NewMetadata(fileData)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(metadata.ProductName()).Should(BeEquivalentTo("cf"))
		})

		It("Should return pivotal-container-service as product name", func() {
			fileData, err := ioutil.ReadFile("fixtures/pks.yml")
			Expect(err).ShouldNot(HaveOccurred())
			metadata, err := generator.NewMetadata(fileData)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(metadata.ProductName()).Should(BeEquivalentTo("pivotal-container-service"))
		})

		It("Should return p-rabbitmq as product name", func() {
			fileData, err := ioutil.ReadFile("fixtures/rabbit-mq.yml")
			Expect(err).ShouldNot(HaveOccurred())
			metadata, err := generator.NewMetadata(fileData)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(metadata.ProductName()).Should(BeEquivalentTo("p-rabbitmq"))
		})

		It("Should return p-healthwatch as product name", func() {
			fileData, err := ioutil.ReadFile("fixtures/p_healthwatch.yml")
			Expect(err).ShouldNot(HaveOccurred())
			metadata, err := generator.NewMetadata(fileData)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(metadata.ProductName()).Should(BeEquivalentTo("p-healthwatch"))
		})

		It("Should return p-isolation-segment as product name", func() {
			fileData, err := ioutil.ReadFile("fixtures/iso-segment.yml")
			Expect(err).ShouldNot(HaveOccurred())
			metadata, err := generator.NewMetadata(fileData)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(metadata.ProductName()).Should(BeEquivalentTo("p-isolation-segment"))
		})

		It("Should return p-isolation-segment-new-seg as product name", func() {
			fileData, err := ioutil.ReadFile("fixtures/iso-segment-replicator.yml")
			Expect(err).ShouldNot(HaveOccurred())
			metadata, err := generator.NewMetadata(fileData)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(metadata.ProductName()).Should(BeEquivalentTo("p-isolation-segment-new-seg"))
		})
	})

	table.DescribeTable("ProductVersion tile metadata fixture tests", func(fixtureFilepath string, expectedVersion string) {
		fileData, err := ioutil.ReadFile(fixtureFilepath)
		Expect(err).ShouldNot(HaveOccurred())
		metadata, err := generator.NewMetadata(fileData)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(metadata.ProductVersion()).Should(BeEquivalentTo(expectedVersion))
	}, table.Entry("PAS", "fixtures/pas.yml", "2.1.3"),
		table.Entry("healthwatch", "fixtures/p_healthwatch.yml", "1.2.1h"),
		table.Entry("iso-segment", "fixtures/iso-segment.yml", "2.2.4"),
		table.Entry("replicated iso-segment", "fixtures/iso-segment-replicator.yml", "2.2.4"),
	)
})
