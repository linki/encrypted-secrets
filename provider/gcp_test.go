package provider

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type GCPSuite struct {
	suite.Suite
}

func (suite *GCPSuite) TestExpandKeyID() {
	for _, tc := range []struct {
		givenKeyID    string
		expandedKeyID string
	}{
		// fully qualified key ID returns the same value
		{
			"projects/my-project/locations/my-region/keyRings/my-keyring/cryptoKeys/my-key",
			"projects/my-project/locations/my-region/keyRings/my-keyring/cryptoKeys/my-key",
		},
		// missing project adds default project
		{
			"locations/my-region/keyRings/my-keyring/cryptoKeys/my-key",
			"projects/default-project/locations/my-region/keyRings/my-keyring/cryptoKeys/my-key",
		},
		// missing region adds default region
		{
			"projects/my-project/keyRings/my-keyring/cryptoKeys/my-key",
			"projects/my-project/locations/default-region/keyRings/my-keyring/cryptoKeys/my-key",
		},
		// missing project and region add default project and region
		{
			"keyRings/my-keyring/cryptoKeys/my-key",
			"projects/default-project/locations/default-region/keyRings/my-keyring/cryptoKeys/my-key",
		},
	} {
		expandedKeyID := expandKeyID(tc.givenKeyID, defaultProject, defaultRegion)
		suite.Equal(tc.expandedKeyID, expandedKeyID)
	}
}

func TestGCPSuite(t *testing.T) {
	suite.Run(t, new(GCPSuite))
}
