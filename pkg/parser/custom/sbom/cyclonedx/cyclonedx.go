package cyclonedx

import (
	"bufio"
	"os"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/google/osv-scanner/pkg/lockfile"
	"github.com/safedep/dry/utils"
	"github.com/safedep/vet/pkg/common/logger"
	"github.com/safedep/vet/pkg/parser/custom/packagefile"
)

func Parse(pathToLockfile string) ([]lockfile.PackageDetails, error) {
	details := []lockfile.PackageDetails{}

	bom := cdx.NewBOM()
	logger.Infof("Starting SBOM decoding...")

	file, err := os.Open(pathToLockfile)
	if err != nil {
		logger.Debugf("Error in Decoding the SBOM file %v", err)
		return nil, err
	}

	defer file.Close()

	sbom_content := bufio.NewReader(file)
	decoder := cdx.NewBOMDecoder(sbom_content, cdx.BOMFileFormatJSON)
	if err = decoder.Decode(bom); err != nil {
		logger.Debugf("Error in Decoding the SBOM file %v", err)
		return nil, err
	}

	// Components is a pointer array and it can be empty
	components := utils.SafelyGetValue(bom.Components)
	for _, comp := range components {
		if d, err := convertSbomComponent2LPD(&comp); err != nil {
			logger.Debugf("Failed converting sbom to lockfile component: %v", err)
		} else {
			details = append(details, *d)
		}
	}

	logger.Debugf("Found number of packages %d", len(details))
	return details, nil
}

func convertSbomComponent2LPD(comp *cdx.Component) (*lockfile.PackageDetails, error) {

	pd, err := packagefile.ParsePackageFromPurl(comp.PackageURL)
	if err != nil {
		return nil, err
	}
	pd.CycloneDxRef = comp
	return pd.Convert2LockfilePackageDetails(), nil
}