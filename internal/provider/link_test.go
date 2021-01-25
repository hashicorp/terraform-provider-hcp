package provider

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	sharedmodels "github.com/hashicorp/cloud-sdk-go/clients/cloud-shared/v1/models"
	"github.com/stretchr/testify/require"
)

func Test_linkURL(t *testing.T) {
	baseLink := &sharedmodels.HashicorpCloudLocationLink{
		Type: "hvn",
		ID:   "test-hvn",
		Location: &sharedmodels.HashicorpCloudLocationLocation{
			OrganizationID: uuid.New().String(),
			ProjectID:      uuid.New().String(),
		},
	}

	t.Run("valid ID", func(t *testing.T) {
		l := *baseLink

		urn, err := linkURL(&l)
		require.NoError(t, err)

		expected := fmt.Sprintf("/organization/%s/project/%s/%s/%s",
			l.Location.OrganizationID,
			l.Location.ProjectID,
			l.Type,
			l.ID)
		require.Equal(t, expected, urn)
	})

	t.Run("missing organization ID", func(t *testing.T) {
		l := *baseLink
		l.Location.OrganizationID = ""

		_, err := linkURL(&l)
		require.Error(t, err)
	})

	t.Run("missing project ID", func(t *testing.T) {
		l := *baseLink
		l.Location.ProjectID = ""

		_, err := linkURL(&l)
		require.Error(t, err)
	})

	t.Run("missing resource type", func(t *testing.T) {
		l := *baseLink
		l.Type = ""

		_, err := linkURL(&l)
		require.Error(t, err)
	})

	t.Run("missing resource ID", func(t *testing.T) {
		l := *baseLink
		l.ID = ""

		_, err := linkURL(&l)
		require.Error(t, err)
	})

	t.Run("missing Location", func(t *testing.T) {
		l := *baseLink
		l.Location = nil

		_, err := linkURL(&l)
		require.Error(t, err)
	})
}

func Test_parseLinkURL(t *testing.T) {
	svcType := "hvn"
	id := "test-hvn"
	orgID := uuid.New().String()
	projID := uuid.New().String()

	t.Run("valid URL", func(t *testing.T) {
		urn := fmt.Sprintf("/organization/%s/project/%s/%s/%s",
			orgID,
			projID,
			svcType,
			id)

		l, err := parseLinkURL(urn)
		require.NoError(t, err)

		require.Equal(t, orgID, l.Location.OrganizationID)
		require.Equal(t, projID, l.Location.ProjectID)
		require.Equal(t, svcType, l.Type)
		require.Equal(t, id, l.ID)
	})

	t.Run("missing organization ID", func(t *testing.T) {
		urn := fmt.Sprintf("/organization/%s/project/%s/%s/%s",
			"",
			projID,
			svcType,
			id)

		_, err := parseLinkURL(urn)
		require.Error(t, err)
	})

	t.Run("missing project ID", func(t *testing.T) {
		urn := fmt.Sprintf("/organization/%s/project/%s/%s/%s",
			orgID,
			"",
			svcType,
			id)

		_, err := parseLinkURL(urn)
		require.Error(t, err)
	})

	t.Run("missing resource type", func(t *testing.T) {
		urn := fmt.Sprintf("/organization/%s/project/%s/%s/%s",
			orgID,
			projID,
			"",
			id)

		_, err := parseLinkURL(urn)
		require.Error(t, err)
	})

	t.Run("missing resource id", func(t *testing.T) {
		urn := fmt.Sprintf("/organization/%s/project/%s/%s/%s",
			orgID,
			projID,
			svcType,
			"")

		_, err := parseLinkURL(urn)
		require.Error(t, err)
	})

	t.Run("missing a field", func(t *testing.T) {
		urn := fmt.Sprintf("/project/%s/%s/%s",
			projID,
			svcType,
			id)

		_, err := parseLinkURL(urn)
		require.Error(t, err)
	})

	t.Run("too many fields", func(t *testing.T) {
		urn := fmt.Sprintf("/extra/value/organization/%s/project/%s/%s/%s",
			orgID,
			projID,
			svcType,
			id)

		_, err := parseLinkURL(urn)
		require.Error(t, err)
	})
}
