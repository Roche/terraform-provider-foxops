package client

import (
	"context"
	"encoding/json"
	e "errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	client_v1 "github.com/Roche/terraform-provider-foxops/internal/client/gen"
	"github.com/Roche/terraform-provider-foxops/internal/helpers"
	"github.com/Roche/terraform-provider-foxops/internal/provider"
	"github.com/deepmap/oapi-codegen/pkg/securityprovider"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
	"github.com/pkg/errors"
)

//go:generate oapi-codegen -generate types,client -package client_v1 -o ./gen/client.gen.go ./openapi.json
//go:generate mockgen -destination ./mocks/http_round_tripper_mock.go -package client_mocks "net/http" RoundTripper

type client struct {
	impl client_v1.ClientInterface
}

type clientOptions struct {
	Transport http.RoundTripper
}

type ClientOption interface {
	apply(opts *clientOptions)
}

type clientTransportOption struct {
	transport http.RoundTripper
}

func ClientTransport(transport http.RoundTripper) clientTransportOption {
	return clientTransportOption{transport}
}

func (o clientTransportOption) apply(opts *clientOptions) {
	opts.Transport = o.transport
}

func New(
	endpoint provider.ClientEndpoint,
	token provider.ClientToken,
	version provider.Version,
	options ...ClientOption,
) provider.FoxopsClient {
	opts := &clientOptions{
		Transport: http.DefaultTransport,
	}

	for _, opt := range options {
		opt.apply(opts)
	}

	provider, err := securityprovider.NewSecurityProviderBearerToken(string(token))
	if err != nil {
		panic(err)
	}

	c, err := client_v1.NewClient(
		string(endpoint),
		client_v1.WithRequestEditorFn(provider.Intercept),
	)
	if err != nil {
		panic(err)
	}

	retryableHttpClient := retryablehttp.NewClient()
	retryableHttpClient.HTTPClient = &http.Client{
		Transport: helpers.NewTransport(string(version), logging.NewLoggingHTTPTransport(opts.Transport)),
		Timeout:   5 * time.Minute,
	}
	c.Client = retryableHttpClient.StandardClient()

	return &client{c}
}

func (c *client) checkResponseStatus(_ context.Context, expected int, resp *http.Response) (err error) {
	if resp.StatusCode != expected {
		var apiError client_v1.ApiError
		var body []byte
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			err = errors.New("failed to read response body")
			return
		}
		err = json.Unmarshal(body, &apiError)
		if err != nil {
			err = errors.Wrap(err, "failed to decode error message")
		} else {
			err = errors.New(apiError.Message)
		}
		err = errors.Wrapf(
			err,
			"unexpected status code %d, wants %d",
			resp.StatusCode,
			expected,
		)
	}
	return
}

func (c *client) GetIncarnation(ctx context.Context, id provider.IncarnationId) (inc provider.Incarnation, err error) {
	var resp *http.Response
	idInt, err := strconv.Atoi(string(id))
	if err != nil {
		err = errors.WithStack(err)
		return
	}

	resp, err = c.impl.ReadIncarnationApiIncarnationsIncarnationIdGet(ctx, idInt)
	if err != nil {
		err = errors.WithStack(err)
		return
	}

	err = errors.WithStack(c.checkResponseStatus(ctx, http.StatusOK, resp))
	if err != nil {
		return
	}

	inc, err = mapIncarnation(resp.Body)

	return
}

func (c *client) GetIncarnationWithMergeRequestStatus(
	ctx context.Context,
	id provider.IncarnationId,
	status string,
) (inc provider.Incarnation, err error) {
	for inc, err = c.GetIncarnation(ctx, id); err == nil; inc, err = c.GetIncarnation(ctx, id) {
		if inc.MergeRequestId == nil {
			return
		}
		if inc.MergeRequestStatus != nil && *inc.MergeRequestStatus == status {
			return
		}
		time.Sleep(time.Second)
	}
	return
}

func (c *client) CreateIncarnation(
	ctx context.Context,
	req provider.CreateIncarnationRequest,
) (inc provider.Incarnation, err error) {
	var resp *http.Response
	body := client_v1.CreateIncarnationApiIncarnationsPostJSONRequestBody{
		IncarnationRepository:     req.IncarnationRepository,
		TargetDirectory:           req.TargetDirectory,
		TemplateData:              map[string]client_v1.DesiredIncarnationState_TemplateData_AdditionalProperties{},
		TemplateRepository:        req.TemplateRepository,
		TemplateRepositoryVersion: req.TemplateRepositoryVersion,
	}

	for key, ivalue := range req.TemplateData {
		data := &client_v1.DesiredIncarnationState_TemplateData_AdditionalProperties{}
		if value, ok := ivalue.(client_v1.DesiredIncarnationStateTemplateData0); ok {
			err = e.Join(err, data.FromDesiredIncarnationStateTemplateData0(value))
		} else if value, ok := ivalue.(client_v1.DesiredIncarnationStateTemplateData1); ok {
			err = e.Join(err, data.FromDesiredIncarnationStateTemplateData1(value))
		} else if value, ok := ivalue.(client_v1.DesiredIncarnationStateTemplateData2); ok {
			err = e.Join(err, data.FromDesiredIncarnationStateTemplateData2(value))
		}
		body.TemplateData[key] = *data
	}
	if err != nil {
		err = errors.WithStack(err)
		return
	}

	resp, err = c.impl.CreateIncarnationApiIncarnationsPost(
		ctx,
		&client_v1.CreateIncarnationApiIncarnationsPostParams{},
		body,
	)
	if err != nil {
		err = errors.WithStack(err)
		return
	}

	err = errors.WithStack(c.checkResponseStatus(ctx, http.StatusCreated, resp))
	if err != nil {
		return
	}

	inc, err = mapIncarnation(resp.Body)

	return
}

func (c *client) UpdateIncarnation(
	ctx context.Context,
	id provider.IncarnationId,
	req provider.UpdateIncarnationRequest,
) (inc provider.Incarnation, err error) {
	var resp *http.Response
	body := client_v1.UpdateIncarnationApiIncarnationsIncarnationIdPutJSONRequestBody{
		Automerge:                 req.AutoMerge,
		TemplateData:              &map[string]client_v1.DesiredIncarnationStatePatch_TemplateData_AdditionalProperties{},
		TemplateRepositoryVersion: &req.TemplateRepositoryVersion,
	}

	for key, ivalue := range req.TemplateData {
		data := &client_v1.DesiredIncarnationStatePatch_TemplateData_AdditionalProperties{}
		if value, ok := ivalue.(client_v1.DesiredIncarnationStateTemplateData0); ok {
			err = e.Join(err, data.FromDesiredIncarnationStatePatchTemplateData0(value))
		} else if value, ok := ivalue.(client_v1.DesiredIncarnationStateTemplateData1); ok {
			err = e.Join(err, data.FromDesiredIncarnationStatePatchTemplateData1(value))
		} else if value, ok := ivalue.(client_v1.DesiredIncarnationStateTemplateData2); ok {
			err = e.Join(err, data.FromDesiredIncarnationStatePatchTemplateData2(value))
		}
		(*body.TemplateData)[key] = *data
	}
	if err != nil {
		err = errors.WithStack(err)
		return
	}

	idInt, err := strconv.Atoi(string(id))
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	resp, err = c.impl.UpdateIncarnationApiIncarnationsIncarnationIdPut(
		ctx,
		idInt,
		body,
	)
	if err != nil {
		err = errors.WithStack(err)
		return
	}

	err = errors.WithStack(c.checkResponseStatus(ctx, http.StatusOK, resp))
	if err != nil {
		return
	}

	inc, err = mapIncarnation(resp.Body)

	return
}

func (c *client) DeleteIncarnation(
	ctx context.Context,
	id provider.IncarnationId,
) (err error) {
	var resp *http.Response
	idInt, err := strconv.Atoi(string(id))
	if err != nil {
		err = errors.WithStack(err)
		return
	}

	resp, err = c.impl.DeleteIncarnationApiIncarnationsIncarnationIdDelete(
		ctx,
		idInt,
	)
	if err != nil {
		err = errors.WithStack(err)
		return
	}

	err = errors.WithStack(c.checkResponseStatus(ctx, http.StatusNoContent, resp))
	if err != nil {
		return
	}
	return
}

func mapIncarnation(body io.Reader) (inc provider.Incarnation, err error) {
	var data client_v1.IncarnationWithDetails
	err = errors.WithStack(json.NewDecoder(body).Decode(&data))
	if err != nil {
		return
	}

	inc = provider.Incarnation{
		Id:                        provider.IncarnationId(fmt.Sprintf("%d", data.Id)),
		IncarnationRepository:     data.IncarnationRepository,
		TargetDirectory:           data.TargetDirectory,
		TemplateData:              make(map[string]interface{}),
		TemplateRepository:        *data.TemplateRepository,
		TemplateRepositoryVersion: *data.TemplateRepositoryVersion,
		MergeRequestUrl:           data.MergeRequestUrl,
		CommitSha:                 data.CommitSha,
		CommitUrl:                 data.CommitUrl,
		MergeRequestId:            data.MergeRequestId,
	}

	if data.MergeRequestStatus != nil {
		status, ok := (*data.MergeRequestStatus).(string)
		if ok {
			inc.MergeRequestStatus = &status
		}
	}

	if data.TemplateData != nil {
		var templateDatum interface{}
		var newErr error
		for key, value := range *data.TemplateData {
			if templateDatum, newErr = value.AsIncarnationWithDetailsTemplateData0(); newErr == nil {
				inc.TemplateData[key] = templateDatum
			} else if templateDatum, newErr = value.AsIncarnationWithDetailsTemplateData1(); newErr == nil {
				inc.TemplateData[key] = templateDatum
			} else if templateDatum, newErr = value.AsIncarnationWithDetailsTemplateData2(); newErr == nil {
				inc.TemplateData[key] = templateDatum
			}
			err = e.Join(err, newErr)
		}
	}

	return
}
