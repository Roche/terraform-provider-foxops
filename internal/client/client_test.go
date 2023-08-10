package client_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/Roche/terraform-provider-foxops/internal/client"
	client_v1 "github.com/Roche/terraform-provider-foxops/internal/client/gen"
	client_mocks "github.com/Roche/terraform-provider-foxops/internal/client/mocks"
	"github.com/Roche/terraform-provider-foxops/internal/helpers"
	"github.com/Roche/terraform-provider-foxops/internal/provider"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type clientTestSetup struct {
	MockRoundTripper    *client_mocks.MockRoundTripper
	Client              provider.FoxopsClient
	AuthorizationHeader client_mocks.RequestMatcherOption
}

func setupClientTest(t *testing.T) *clientTestSetup {
	ctrl := gomock.NewController(t)

	mockRoundTripper := client_mocks.NewMockRoundTripper(ctrl)
	token := "dev-token"
	c := client.New(
		provider.ClientEndpoint("http://localhost"),
		provider.ClientToken(token),
		"testing",
		client.ClientTransport(mockRoundTripper),
	)

	return &clientTestSetup{
		Client:              c,
		MockRoundTripper:    mockRoundTripper,
		AuthorizationHeader: client_mocks.RequestHeader("authorization", fmt.Sprintf("Bearer %s", token)),
	}
}

func TestClient_GetIncarnation_ShouldSucceedWhenReceivingOk(t *testing.T) {
	setup := setupClientTest(t)

	ctx := context.Background()

	id := 1234

	want := provider.Incarnation{
		Id:                        provider.IncarnationId(fmt.Sprintf("%d", id)),
		IncarnationRepository:     "inc/repo",
		TemplateRepository:        "template/repo",
		TemplateRepositoryVersion: "template/repo/version",
		TargetDirectory:           ".",
		TemplateData:              map[string]interface{}{},
		CommitSha:                 "12345678",
		CommitUrl:                 "template/repo/commit",
	}

	body, err := json.Marshal(
		client_v1.IncarnationWithDetails{
			Id:                        id,
			IncarnationRepository:     want.IncarnationRepository,
			TemplateRepository:        &want.TemplateRepository,
			TemplateRepositoryVersion: &want.TemplateRepositoryVersion,
			TargetDirectory:           want.TargetDirectory,
			TemplateData:              &map[string]client_v1.IncarnationWithDetails_TemplateData_AdditionalProperties{},
			CommitSha:                 want.CommitSha,
			CommitUrl:                 want.CommitUrl,
		},
	)
	require.NoError(t, err)

	response := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBuffer(body)),
		Header:     make(http.Header),
	}

	setup.MockRoundTripper.EXPECT().
		RoundTrip(
			client_mocks.NewRequestMatcher(
				client_mocks.RequestMethod(http.MethodGet),
				client_mocks.RequestPathf("/api/incarnations/%d", id),
				setup.AuthorizationHeader,
			),
		).
		Return(response, nil)

	got, err := setup.Client.GetIncarnation(ctx, want.Id)

	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestClient_GetIncarnationWithMergeRequestStatus_ShouldSucceedWhenReceivingOkAndNoMergeRequestId(
	t *testing.T,
) {
	setup := setupClientTest(t)

	ctx := context.Background()

	id := 1234

	want := provider.Incarnation{
		Id:                        provider.IncarnationId(fmt.Sprintf("%d", id)),
		IncarnationRepository:     "inc/repo",
		TemplateRepository:        "template/repo",
		TemplateRepositoryVersion: "template/repo/version",
		TargetDirectory:           ".",
		TemplateData:              map[string]interface{}{},
		CommitSha:                 "12345678",
		CommitUrl:                 "template/repo/commit",
	}

	body, err := json.Marshal(
		client_v1.IncarnationWithDetails{
			Id:                        id,
			IncarnationRepository:     want.IncarnationRepository,
			TemplateRepository:        &want.TemplateRepository,
			TemplateRepositoryVersion: &want.TemplateRepositoryVersion,
			TargetDirectory:           want.TargetDirectory,
			TemplateData:              &map[string]client_v1.IncarnationWithDetails_TemplateData_AdditionalProperties{},
			CommitSha:                 want.CommitSha,
			CommitUrl:                 want.CommitUrl,
		},
	)
	require.NoError(t, err)

	response := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBuffer(body)),
		Header:     make(http.Header),
	}

	setup.MockRoundTripper.EXPECT().
		RoundTrip(
			client_mocks.NewRequestMatcher(
				client_mocks.RequestMethod(http.MethodGet),
				client_mocks.RequestPathf("/api/incarnations/%d", id),
				setup.AuthorizationHeader,
			),
		).
		Return(response, nil)

	got, err := setup.Client.GetIncarnationWithMergeRequestStatus(ctx, want.Id, "some-status")

	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestClient_GetIncarnationWithMergeRequestStatus_ShouldSucceedWhenReceivingOkAndExpectedMergeRequestStatus(
	t *testing.T,
) {
	setup := setupClientTest(t)

	ctx := context.Background()

	id := 1234

	want := provider.Incarnation{
		Id:                        provider.IncarnationId(fmt.Sprintf("%d", id)),
		IncarnationRepository:     "inc/repo",
		TemplateRepository:        "template/repo",
		TemplateRepositoryVersion: "template/repo/version",
		TargetDirectory:           ".",
		TemplateData:              map[string]interface{}{},
		CommitSha:                 "12345678",
		CommitUrl:                 "template/repo/commit",
		MergeRequestId:            helpers.Addr("1234"),
		MergeRequestStatus:        helpers.Addr("merged"),
	}

	body, err := json.Marshal(
		client_v1.IncarnationWithDetails{
			Id:                        id,
			IncarnationRepository:     want.IncarnationRepository,
			TemplateRepository:        &want.TemplateRepository,
			TemplateRepositoryVersion: &want.TemplateRepositoryVersion,
			TargetDirectory:           want.TargetDirectory,
			TemplateData:              &map[string]client_v1.IncarnationWithDetails_TemplateData_AdditionalProperties{},
			CommitSha:                 want.CommitSha,
			CommitUrl:                 want.CommitUrl,
			MergeRequestId:            want.MergeRequestId,
			MergeRequestStatus:        helpers.Addr[interface{}](*want.MergeRequestStatus),
		},
	)
	require.NoError(t, err)

	response := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBuffer(body)),
		Header:     make(http.Header),
	}

	setup.MockRoundTripper.EXPECT().
		RoundTrip(
			client_mocks.NewRequestMatcher(
				client_mocks.RequestMethod(http.MethodGet),
				client_mocks.RequestPathf("/api/incarnations/%d", id),
				setup.AuthorizationHeader,
			),
		).
		Return(response, nil)

	got, err := setup.Client.GetIncarnationWithMergeRequestStatus(ctx, want.Id, "merged")

	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestClient_GetIncarnationWithMergeRequestStatus_ShouldRetryWhenReceivingOkAndUnexpectedMergeRequestStatus(
	t *testing.T,
) {
	setup := setupClientTest(t)

	ctx := context.Background()

	id := 1234

	want := provider.Incarnation{
		Id:                        provider.IncarnationId(fmt.Sprintf("%d", id)),
		IncarnationRepository:     "inc/repo",
		TemplateRepository:        "template/repo",
		TemplateRepositoryVersion: "template/repo/version",
		TargetDirectory:           ".",
		TemplateData:              map[string]interface{}{},
		CommitSha:                 "12345678",
		CommitUrl:                 "template/repo/commit",
		MergeRequestId:            helpers.Addr("1234"),
		MergeRequestStatus:        helpers.Addr("merged"),
	}

	body1, err := json.Marshal(
		client_v1.IncarnationWithDetails{
			Id:                        id,
			IncarnationRepository:     want.IncarnationRepository,
			TemplateRepository:        &want.TemplateRepository,
			TemplateRepositoryVersion: &want.TemplateRepositoryVersion,
			TargetDirectory:           want.TargetDirectory,
			TemplateData:              &map[string]client_v1.IncarnationWithDetails_TemplateData_AdditionalProperties{},
			CommitSha:                 want.CommitSha,
			CommitUrl:                 want.CommitUrl,
			MergeRequestId:            want.MergeRequestId,
			MergeRequestStatus:        helpers.Addr[interface{}]("open"),
		},
	)
	require.NoError(t, err)

	body2, err := json.Marshal(
		client_v1.IncarnationWithDetails{
			Id:                        id,
			IncarnationRepository:     want.IncarnationRepository,
			TemplateRepository:        &want.TemplateRepository,
			TemplateRepositoryVersion: &want.TemplateRepositoryVersion,
			TargetDirectory:           want.TargetDirectory,
			TemplateData:              &map[string]client_v1.IncarnationWithDetails_TemplateData_AdditionalProperties{},
			CommitSha:                 want.CommitSha,
			CommitUrl:                 want.CommitUrl,
			MergeRequestId:            want.MergeRequestId,
			MergeRequestStatus:        helpers.Addr[interface{}](*want.MergeRequestStatus),
		},
	)
	require.NoError(t, err)

	response1 := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBuffer(body1)),
		Header:     make(http.Header),
	}

	setup.MockRoundTripper.EXPECT().
		RoundTrip(
			client_mocks.NewRequestMatcher(
				client_mocks.RequestMethod(http.MethodGet),
				client_mocks.RequestPathf("/api/incarnations/%d", id),
				setup.AuthorizationHeader,
			),
		).
		Return(response1, nil)

	response2 := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBuffer(body2)),
		Header:     make(http.Header),
	}

	setup.MockRoundTripper.EXPECT().
		RoundTrip(
			client_mocks.NewRequestMatcher(
				client_mocks.RequestMethod(http.MethodGet),
				client_mocks.RequestPathf("/api/incarnations/%d", id),
				setup.AuthorizationHeader,
			),
		).
		Return(response2, nil)

	got, err := setup.Client.GetIncarnationWithMergeRequestStatus(ctx, want.Id, "merged")

	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestClient_CreateIncarnation_ShouldSucceedWhenReceivingCreated(
	t *testing.T,
) {
	setup := setupClientTest(t)

	ctx := context.Background()

	id := 1234

	want := provider.Incarnation{
		Id:                        provider.IncarnationId(fmt.Sprintf("%d", id)),
		IncarnationRepository:     "inc/repo",
		TemplateRepository:        "template/repo",
		TemplateRepositoryVersion: "template/repo/version",
		TargetDirectory:           ".",
		TemplateData:              map[string]interface{}{},
		CommitSha:                 "12345678",
		CommitUrl:                 "template/repo/commit",
		MergeRequestId:            helpers.Addr("1234"),
		MergeRequestStatus:        helpers.Addr("merged"),
	}

	req := provider.CreateIncarnationRequest{
		IncarnationRepository: want.IncarnationRepository,
		TargetDirectory:       &want.TargetDirectory,
		TemplateRepository:    want.TemplateRepository,
		UpdateIncarnationRequest: provider.UpdateIncarnationRequest{
			TemplateData:              want.TemplateData,
			TemplateRepositoryVersion: want.TemplateRepositoryVersion,
		},
	}

	body, err := json.Marshal(
		client_v1.IncarnationWithDetails{
			Id:                        id,
			IncarnationRepository:     want.IncarnationRepository,
			TemplateRepository:        &want.TemplateRepository,
			TemplateRepositoryVersion: &want.TemplateRepositoryVersion,
			TargetDirectory:           want.TargetDirectory,
			TemplateData:              &map[string]client_v1.IncarnationWithDetails_TemplateData_AdditionalProperties{},
			CommitSha:                 want.CommitSha,
			CommitUrl:                 want.CommitUrl,
			MergeRequestId:            want.MergeRequestId,
			MergeRequestStatus:        helpers.Addr[interface{}](*want.MergeRequestStatus),
		},
	)
	require.NoError(t, err)

	response := &http.Response{
		StatusCode: http.StatusCreated,
		Body:       io.NopCloser(bytes.NewBuffer(body)),
		Header:     make(http.Header),
	}

	setup.MockRoundTripper.EXPECT().
		RoundTrip(
			client_mocks.NewRequestMatcher(
				client_mocks.RequestMethod(http.MethodPost),
				client_mocks.RequestPath("/api/incarnations"),
				setup.AuthorizationHeader,
			),
		).
		Return(response, nil)

	got, err := setup.Client.CreateIncarnation(ctx, req)

	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestClient_UpdateIncarnation_ShouldSucceedWhenReceivingOk(
	t *testing.T,
) {
	setup := setupClientTest(t)

	ctx := context.Background()

	id := 1234

	want := provider.Incarnation{
		Id:                        provider.IncarnationId(fmt.Sprintf("%d", id)),
		IncarnationRepository:     "inc/repo",
		TemplateRepository:        "template/repo",
		TemplateRepositoryVersion: "template/repo/version",
		TargetDirectory:           ".",
		TemplateData:              map[string]interface{}{},
		CommitSha:                 "12345678",
		CommitUrl:                 "template/repo/commit",
		MergeRequestId:            helpers.Addr("1234"),
		MergeRequestStatus:        helpers.Addr("merged"),
	}

	req := provider.UpdateIncarnationRequest{
		TemplateData:              want.TemplateData,
		TemplateRepositoryVersion: want.TemplateRepositoryVersion,
	}

	body, err := json.Marshal(
		client_v1.IncarnationWithDetails{
			Id:                        id,
			IncarnationRepository:     want.IncarnationRepository,
			TemplateRepository:        &want.TemplateRepository,
			TemplateRepositoryVersion: &want.TemplateRepositoryVersion,
			TargetDirectory:           want.TargetDirectory,
			TemplateData:              &map[string]client_v1.IncarnationWithDetails_TemplateData_AdditionalProperties{},
			CommitSha:                 want.CommitSha,
			CommitUrl:                 want.CommitUrl,
			MergeRequestId:            want.MergeRequestId,
			MergeRequestStatus:        helpers.Addr[interface{}](*want.MergeRequestStatus),
		},
	)
	require.NoError(t, err)

	response := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBuffer(body)),
		Header:     make(http.Header),
	}

	setup.MockRoundTripper.EXPECT().
		RoundTrip(
			client_mocks.NewRequestMatcher(
				client_mocks.RequestMethod(http.MethodPut),
				client_mocks.RequestPathf("/api/incarnations/%s", want.Id),
				setup.AuthorizationHeader,
			),
		).
		Return(response, nil)

	got, err := setup.Client.UpdateIncarnation(ctx, want.Id, req)

	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestClient_DeleteIncarnation_ShouldSucceedWhenReceivingOk(
	t *testing.T,
) {
	setup := setupClientTest(t)

	ctx := context.Background()

	id := provider.IncarnationId("1234")

	response := &http.Response{
		StatusCode: http.StatusNoContent,
		Body:       io.NopCloser(bytes.NewBuffer([]byte{})),
		Header:     make(http.Header),
	}

	setup.MockRoundTripper.EXPECT().
		RoundTrip(
			client_mocks.NewRequestMatcher(
				client_mocks.RequestMethod(http.MethodDelete),
				client_mocks.RequestPathf("/api/incarnations/%s", id),
				setup.AuthorizationHeader,
			),
		).
		Return(response, nil)

	err := setup.Client.DeleteIncarnation(ctx, id)

	require.NoError(t, err)
}
