package provider

import (
	"context"
)

type IncarnationId string

type Incarnation struct {
	Id                        IncarnationId
	IncarnationRepository     string
	TargetDirectory           string
	TemplateData              map[string]interface{}
	TemplateRepository        string
	TemplateRepositoryVersion string
	MergeRequestUrl           *string
	CommitSha                 string
	CommitUrl                 string
	MergeRequestStatus        *string
	MergeRequestId            *string
}

type UpdateIncarnationRequest struct {
	AutoMerge                 bool
	TemplateData              map[string]interface{}
	TemplateRepositoryVersion string
}

type CreateIncarnationRequest struct {
	UpdateIncarnationRequest
	IncarnationRepository string
	TargetDirectory       *string
	TemplateRepository    string
}

//go:generate mockgen -destination ./mocks/client_mock.go . FoxopsClient
type FoxopsClient interface {
	GetIncarnation(context.Context, IncarnationId) (Incarnation, error)
	GetIncarnationWithMergeRequestStatus(context.Context, IncarnationId, string) (Incarnation, error)
	CreateIncarnation(context.Context, CreateIncarnationRequest) (Incarnation, error)
	UpdateIncarnation(context.Context, IncarnationId, UpdateIncarnationRequest) (Incarnation, error)
	DeleteIncarnation(context.Context, IncarnationId) error
}
