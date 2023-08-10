package provider_test

import (
	"fmt"
	"testing"

	"github.com/Roche/terraform-provider-foxops/internal/helpers"
	"github.com/Roche/terraform-provider-foxops/internal/provider"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestAcc_IncarnationDataSource(t *testing.T) {
	setup := newTestProviderSetup(t)

	incarnation := provider.Incarnation{
		Id:                        provider.IncarnationId("1234"),
		IncarnationRepository:     "inc/repo",
		TemplateRepository:        "template/repo",
		TemplateRepositoryVersion: "template/repo/version",
		TargetDirectory:           ".",
		CommitSha:                 "12345678",
		CommitUrl:                 "template/repo/commit",
		MergeRequestId:            helpers.Addr("1234"),
		MergeRequestStatus:        helpers.Addr("merged"),
		MergeRequestUrl:           helpers.Addr("inc/repo/mr!1234"),
		TemplateData: map[string]interface{}{
			"hello": "World!",
		},
	}

	setup.client.EXPECT().
		GetIncarnation(gomock.Any(), incarnation.Id).
		Return(incarnation, nil).
		Times(5)

	hello, ok := incarnation.TemplateData["hello"].(string)
	require.True(t, ok)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: setup.testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(
					providerConfig+
						`data "foxops_incarnation" "test" {
  id   = "%s"
}`,
					incarnation.Id,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.foxops_incarnation.test", "id", string(incarnation.Id)),
					resource.TestCheckResourceAttr("data.foxops_incarnation.test", "incarnation_repository", incarnation.IncarnationRepository),
					resource.TestCheckResourceAttr("data.foxops_incarnation.test", "template_repository", incarnation.TemplateRepository),
					resource.TestCheckResourceAttr("data.foxops_incarnation.test", "template_repository_version", incarnation.TemplateRepositoryVersion),
					resource.TestCheckResourceAttr("data.foxops_incarnation.test", "target_directory", incarnation.TargetDirectory),
					resource.TestCheckResourceAttr("data.foxops_incarnation.test", "template_data.hello", hello),
					resource.TestCheckResourceAttr("data.foxops_incarnation.test", "commit_sha", incarnation.CommitSha),
					resource.TestCheckResourceAttr("data.foxops_incarnation.test", "commit_url", incarnation.CommitUrl),
					resource.TestCheckResourceAttr("data.foxops_incarnation.test", "merge_request_id", *incarnation.MergeRequestId),
					resource.TestCheckResourceAttr("data.foxops_incarnation.test", "merge_request_status", *incarnation.MergeRequestStatus),
					resource.TestCheckResourceAttr("data.foxops_incarnation.test", "merge_request_url", *incarnation.MergeRequestUrl),
				),
			},
		},
	})
}

func TestAcc_IncarnationDataSource_WithWaitForMRStatus(t *testing.T) {
	setup := newTestProviderSetup(t)

	incarnation := provider.Incarnation{
		Id:                        provider.IncarnationId("1234"),
		IncarnationRepository:     "inc/repo",
		TemplateRepository:        "template/repo",
		TemplateRepositoryVersion: "template/repo/version",
		TargetDirectory:           ".",
		CommitSha:                 "12345678",
		CommitUrl:                 "template/repo/commit",
		MergeRequestId:            helpers.Addr("1234"),
		MergeRequestStatus:        helpers.Addr("merged"),
		MergeRequestUrl:           helpers.Addr("inc/repo/mr!1234"),
		TemplateData: map[string]interface{}{
			"hello": "World!",
		},
	}

	hello, ok := incarnation.TemplateData["hello"].(string)
	require.True(t, ok)

	setup.client.EXPECT().
		GetIncarnationWithMergeRequestStatus(gomock.Any(), incarnation.Id, *incarnation.MergeRequestStatus).
		Return(incarnation, nil).
		Times(5)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: setup.testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(
					providerConfig+
						`data "foxops_incarnation" "test" {
  id   = "%s"
  wait_for_mr_status = {
	status = "merged"
  }
}`,
					incarnation.Id,
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.foxops_incarnation.test", "id", string(incarnation.Id)),
					resource.TestCheckResourceAttr("data.foxops_incarnation.test", "incarnation_repository", incarnation.IncarnationRepository),
					resource.TestCheckResourceAttr("data.foxops_incarnation.test", "template_repository", incarnation.TemplateRepository),
					resource.TestCheckResourceAttr("data.foxops_incarnation.test", "template_repository_version", incarnation.TemplateRepositoryVersion),
					resource.TestCheckResourceAttr("data.foxops_incarnation.test", "target_directory", incarnation.TargetDirectory),
					resource.TestCheckResourceAttr("data.foxops_incarnation.test", "template_data.hello", hello),
					resource.TestCheckResourceAttr("data.foxops_incarnation.test", "commit_sha", incarnation.CommitSha),
					resource.TestCheckResourceAttr("data.foxops_incarnation.test", "commit_url", incarnation.CommitUrl),
					resource.TestCheckResourceAttr("data.foxops_incarnation.test", "merge_request_id", *incarnation.MergeRequestId),
					resource.TestCheckResourceAttr("data.foxops_incarnation.test", "merge_request_status", *incarnation.MergeRequestStatus),
					resource.TestCheckResourceAttr("data.foxops_incarnation.test", "merge_request_url", *incarnation.MergeRequestUrl),
				),
			},
		},
	})
}
