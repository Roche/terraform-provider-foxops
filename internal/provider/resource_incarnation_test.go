package provider_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/Roche/terraform-provider-foxops/internal/helpers"
	"github.com/Roche/terraform-provider-foxops/internal/provider"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/imdario/mergo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func incarnationResourceConfigFactory(name string, req provider.CreateIncarnationRequest) string {
	result := providerConfig
	result += fmt.Sprintf(`resource "foxops_incarnation" "%s" {`, name) + "\n"
	result += fmt.Sprintf(`  incarnation_repository = "%s"`, req.IncarnationRepository) + "\n"
	result += fmt.Sprintf(`  target_directory = "%s"`, *req.TargetDirectory) + "\n"
	result += fmt.Sprintf(`  template_repository = "%s"`, req.TemplateRepository) + "\n"
	result += fmt.Sprintf(`  template_repository_version = "%s"`, req.TemplateRepositoryVersion) + "\n"
	result += `  template_data = {` + "\n"
	for key, value := range req.TemplateData {
		result += fmt.Sprintf(`    %s = "%s"`, key, value) + "\n"
	}
	result += `  }` + "\n"
	result += `}` + "\n"
	return result
}

func TestAccEdgeClusterResource_ShouldCreateOrImportAnIncarnation(t *testing.T) {
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

	req := provider.CreateIncarnationRequest{
		IncarnationRepository: incarnation.IncarnationRepository,
		TargetDirectory:       &incarnation.TargetDirectory,
		TemplateRepository:    incarnation.TemplateRepository,
		UpdateIncarnationRequest: provider.UpdateIncarnationRequest{
			TemplateData:              incarnation.TemplateData,
			TemplateRepositoryVersion: incarnation.TemplateRepositoryVersion,
		},
	}

	setup.client.EXPECT().
		CreateIncarnation(gomock.Any(), req).
		Return(incarnation, nil)

	setup.client.EXPECT().
		GetIncarnation(gomock.Any(), incarnation.Id).
		Return(incarnation, nil).
		Times(2)

	setup.client.EXPECT().
		DeleteIncarnation(gomock.Any(), incarnation.Id).
		Return(nil)

	hello, ok := incarnation.TemplateData["hello"].(string)
	require.True(t, ok)
	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: setup.testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: incarnationResourceConfigFactory("test", req),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("foxops_incarnation.test", "id", string(incarnation.Id)),
					resource.TestCheckResourceAttr("foxops_incarnation.test", "incarnation_repository", incarnation.IncarnationRepository),
					resource.TestCheckResourceAttr("foxops_incarnation.test", "template_repository", incarnation.TemplateRepository),
					resource.TestCheckResourceAttr("foxops_incarnation.test", "template_repository_version", incarnation.TemplateRepositoryVersion),
					resource.TestCheckResourceAttr("foxops_incarnation.test", "target_directory", incarnation.TargetDirectory),
					resource.TestCheckResourceAttr("foxops_incarnation.test", "template_data.hello", hello),
					resource.TestCheckResourceAttr("foxops_incarnation.test", "commit_sha", incarnation.CommitSha),
					resource.TestCheckResourceAttr("foxops_incarnation.test", "commit_url", incarnation.CommitUrl),
					resource.TestCheckResourceAttr("foxops_incarnation.test", "merge_request_id", *incarnation.MergeRequestId),
					resource.TestCheckResourceAttr("foxops_incarnation.test", "merge_request_status", *incarnation.MergeRequestStatus),
					resource.TestCheckResourceAttr("foxops_incarnation.test", "merge_request_url", *incarnation.MergeRequestUrl),
				),
			},
			{
				ResourceName:       "foxops_incarnation.test",
				ImportState:        true,
				ImportStatePersist: false,
				ImportStateVerify:  true,
			},
		},
	})
}

type resourceChangeTestSetup struct {
	Name           string
	Key            string
	Value          interface{}
	ShouldRecreate bool
}

func getField(v interface{}, field string) interface{} {
	r := reflect.ValueOf(v)
	f := reflect.Indirect(r).FieldByName(field)
	return f.Interface()
}

func TestIncarnationResourceChanges(t *testing.T) {

	incarnationResourceChangeTestName := func(key string, shouldRecreate bool) string {
		result := fmt.Sprintf("When%sChanges_ItShould", key)
		if shouldRecreate {
			return result + "RecreateTheResource"
		}
		return result + "UpdateTheResource"

	}
	for _, changeTestSetup := range []resourceChangeTestSetup{
		{
			Key:            "IncarnationRepository",
			Value:          "inc/other-repo",
			ShouldRecreate: true,
		},
		{
			Key:            "TargetDirectory",
			Value:          helpers.Addr("./dir"),
			ShouldRecreate: true,
		},
		{
			Key:            "TemplateData",
			Value:          map[string]interface{}{"hello": "You!"},
			ShouldRecreate: false,
		},
		{
			Key:            "TemplateRepository",
			Value:          "template/other-repo",
			ShouldRecreate: true,
		},
		{
			Key:            "TemplateRepositoryVersion",
			Value:          "template/repo/other-version",
			ShouldRecreate: false,
		},
	} {
		t.Run(
			incarnationResourceChangeTestName(changeTestSetup.Key, changeTestSetup.ShouldRecreate),
			func(t *testing.T) {
				setup := newTestProviderSetup(t)

				incarnation := provider.Incarnation{
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

				changes := map[string]interface{}{changeTestSetup.Key: changeTestSetup.Value}

				id := 0
				req1 := provider.CreateIncarnationRequest{
					IncarnationRepository: incarnation.IncarnationRepository,
					TemplateRepository:    incarnation.TemplateRepository,
					TargetDirectory:       &incarnation.TargetDirectory,
					UpdateIncarnationRequest: provider.UpdateIncarnationRequest{
						TemplateRepositoryVersion: incarnation.TemplateRepositoryVersion,
						TemplateData:              incarnation.TemplateData,
					},
				}
				req2 := provider.CreateIncarnationRequest{}

				require.NoError(t, mergo.Map(&req2, changes, mergo.WithOverride))
				require.NoError(t, mergo.Merge(&req2, req1))

				createCallCount := 1
				if changeTestSetup.ShouldRecreate {
					createCallCount += 1
				}

				updateCallCount := 0
				if !changeTestSetup.ShouldRecreate {
					updateCallCount += 1
				}

				id1 := "0001"
				id2 := id1
				if changeTestSetup.ShouldRecreate {
					id2 = "0002"
				}

				setup.client.EXPECT().
					CreateIncarnation(
						gomock.Any(),
						gomock.Any(),
					).
					DoAndReturn(
						func(_ context.Context, req provider.CreateIncarnationRequest) (ec provider.Incarnation, err error) {
							id += 1
							tmp := map[string]interface{}{
								"id": provider.IncarnationId(fmt.Sprintf("%04d", id)),
							}
							err = mergo.Map(&tmp, req)
							if err != nil {
								return
							}
							err = mergo.Map(&incarnation, tmp, mergo.WithOverride)
							if err != nil {
								return
							}
							return incarnation, nil
						},
					).
					Times(createCallCount)

				setup.client.EXPECT().
					UpdateIncarnation(
						gomock.Any(),
						gomock.Any(),
						gomock.Any(),
					).
					DoAndReturn(
						func(
							_ context.Context,
							id provider.IncarnationId,
							req provider.UpdateIncarnationRequest,
						) (ec provider.Incarnation, err error) {
							assert.Equal(t, id, incarnation.Id)
							assert.Equal(t, getField(req, changeTestSetup.Key), changeTestSetup.Value)
							tmp := map[string]interface{}{}
							err = mergo.Map(&tmp, req)
							if err != nil {
								return
							}
							err = mergo.Map(&incarnation, tmp, mergo.WithOverride)
							if err != nil {
								return
							}
							return incarnation, nil
						},
					).
					Times(updateCallCount)

				setup.client.EXPECT().
					GetIncarnation(
						gomock.Any(),
						gomock.Any(),
					).
					DoAndReturn(
						func(_ context.Context, id provider.IncarnationId) (provider.Incarnation, error) {
							assert.Equal(t, id, incarnation.Id)
							return incarnation, nil
						},
					).
					Times(3 + updateCallCount)

				setup.client.EXPECT().
					DeleteIncarnation(
						gomock.Any(),
						gomock.Any(),
					).
					DoAndReturn(
						func(_ context.Context, id provider.IncarnationId) error {
							assert.Equal(t, id, incarnation.Id)
							return nil
						},
					).
					Times(createCallCount)

				hello1, ok1 := req1.TemplateData["hello"].(string)
				hello2, ok2 := req2.TemplateData["hello"].(string)
				require.True(t, ok1 && ok2)

				resource.Test(t, resource.TestCase{
					IsUnitTest:               true,
					ProtoV6ProviderFactories: setup.testAccProtoV6ProviderFactories,
					Steps: []resource.TestStep{
						{
							Config: incarnationResourceConfigFactory("test", req1),
							Check: resource.ComposeTestCheckFunc(
								resource.TestCheckResourceAttr("foxops_incarnation.test", "id", id1),
								resource.TestCheckResourceAttr("foxops_incarnation.test", "incarnation_repository", req1.IncarnationRepository),
								resource.TestCheckResourceAttr("foxops_incarnation.test", "target_directory", *req1.TargetDirectory),
								resource.TestCheckResourceAttr("foxops_incarnation.test", "template_data.hello", hello1),
								resource.TestCheckResourceAttr("foxops_incarnation.test", "template_repository", req1.TemplateRepository),
								resource.TestCheckResourceAttr("foxops_incarnation.test", "template_repository_version", req1.TemplateRepositoryVersion),
							),
						},
						{
							Config: incarnationResourceConfigFactory("test", req2),
							Check: resource.ComposeTestCheckFunc(
								resource.TestCheckResourceAttr("foxops_incarnation.test", "id", id2),
								resource.TestCheckResourceAttr("foxops_incarnation.test", "incarnation_repository", req2.IncarnationRepository),
								resource.TestCheckResourceAttr("foxops_incarnation.test", "target_directory", *req2.TargetDirectory),
								resource.TestCheckResourceAttr("foxops_incarnation.test", "template_data.hello", hello2),
								resource.TestCheckResourceAttr("foxops_incarnation.test", "template_repository", req2.TemplateRepository),
								resource.TestCheckResourceAttr("foxops_incarnation.test", "template_repository_version", req2.TemplateRepositoryVersion),
							),
						},
					},
				})
			},
		)
	}
}
