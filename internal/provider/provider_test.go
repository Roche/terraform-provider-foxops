package provider_test

import (
	"testing"

	"github.com/Roche/terraform-provider-foxops/internal/provider"
	mock_provider "github.com/Roche/terraform-provider-foxops/internal/provider/mocks"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"go.uber.org/mock/gomock"
)

const (
	providerConfig = `
provider "foxops" {
	endpoint = "http://localhost:9876"
	token = "fake-token"
}
`
)

type testProviderSetup struct {
	client                          *mock_provider.MockFoxopsClient
	testAccProtoV6ProviderFactories map[string]func() (tfprotov6.ProviderServer, error)
}

func newTestProviderSetup(
	t *testing.T,
) testProviderSetup {
	ctrl := gomock.NewController(t)

	client := mock_provider.NewMockFoxopsClient(ctrl)

	return testProviderSetup{
		client: client,
		testAccProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"foxops": providerserver.NewProtocol6WithError(
				provider.New(
					"test",
					func(provider.ClientEndpoint, provider.ClientToken, provider.Version) provider.FoxopsClient {
						return client
					},
					[]func() datasource.DataSource{provider.NewIncarnationDataSource},
					[]func() resource.Resource{provider.NewIncarnationResource},
				)(),
			),
		},
	}
}
