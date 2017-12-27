package commands

import (
	"errors"
	"fmt"

	"github.com/pivotal-cf/jhanda"
	"github.com/pivotal-cf/om/api"
)

type ImportInstallation struct {
	multipart                        multipart
	logger                           logger
	installationAssetImporterService installationAssetImporterService
	setupService                     setupService
	Options                          struct {
		Installation    string `short:"i"  long:"installation"  description:"path to installation."`
		Passphrase      string `short:"dp" long:"decryption-passphrase" description:"passphrase for Ops Manager to decrypt the installation"`
		PollingInterval int    `short:"pi" long:"polling-interval" description:"interval (in seconds) at which to print status" default:"1"`
	}
}

//go:generate counterfeiter -o ./fakes/installation_asset_importer_service.go --fake-name InstallationAssetImporterService . installationAssetImporterService
type installationAssetImporterService interface {
	Import(api.ImportInstallationInput) error
}

func NewImportInstallation(multipart multipart, installationAssetImporterService installationAssetImporterService, setupService setupService, logger logger) ImportInstallation {
	return ImportInstallation{
		multipart: multipart,
		logger:    logger,
		installationAssetImporterService: installationAssetImporterService,
		setupService:                     setupService,
	}
}

func (ii ImportInstallation) Usage() jhanda.Usage {
	return jhanda.Usage{
		Description:      "This unauthenticated command attempts to import an installation to the Ops Manager targeted.",
		ShortDescription: "imports a given installation to the Ops Manager targeted",
		Flags:            ii.Options,
	}
}

func (ii ImportInstallation) Execute(args []string) error {
	_, err := jhanda.Parse(&ii.Options, args)
	if err != nil {
		return fmt.Errorf("could not parse import-installation flags: %s", err)
	}

	if ii.Options.Passphrase == "" {
		return errors.New("could not parse import-installation flags: decryption passphrase not provided")
	}

	ensureAvailabilityOutput, err := ii.setupService.EnsureAvailability(api.EnsureAvailabilityInput{})
	if err != nil {
		return fmt.Errorf("could not check Ops Manager status: %s", err)
	}

	if ensureAvailabilityOutput.Status != api.EnsureAvailabilityStatusUnstarted {
		return errors.New("cannot import installation to an Ops Manager that is already configured")
	}

	ii.logger.Printf("processing installation")

	err = ii.multipart.AddFile("installation[file]", ii.Options.Installation)
	if err != nil {
		return fmt.Errorf("failed to load installation: %s", err)
	}

	err = ii.multipart.AddField("passphrase", ii.Options.Passphrase)
	if err != nil {
		return fmt.Errorf("failed to insert passphrase: %s", err)
	}

	submission, err := ii.multipart.Finalize()
	if err != nil {
		return fmt.Errorf("failed to create multipart form: %s", err)
	}

	ii.logger.Printf("beginning installation import to Ops Manager")

	err = ii.installationAssetImporterService.Import(api.ImportInstallationInput{
		ContentLength:   submission.Length,
		Installation:    submission.Content,
		ContentType:     submission.ContentType,
		PollingInterval: ii.Options.PollingInterval,
	})
	if err != nil {
		return fmt.Errorf("failed to import installation: %s", err)
	}

	ii.logger.Printf("waiting for import to complete...")
	for ensureAvailabilityOutput.Status != api.EnsureAvailabilityStatusComplete {
		ensureAvailabilityOutput, err = ii.setupService.EnsureAvailability(api.EnsureAvailabilityInput{})
		if err != nil {
			return fmt.Errorf("could not check Ops Manager Status: %s", err)
		}
	}

	ii.logger.Printf("finished import")

	return nil
}
