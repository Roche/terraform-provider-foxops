package provider

import (
	"context"
	"errors"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func getIncarnation(
	ctx context.Context,
	client FoxopsClient,
	id IncarnationId,
	waitForStatus *waitForStatusMRModel,
) (inc Incarnation, err error, diags diag.Diagnostics) {
	var clientErr error
	if waitForStatus == nil {
		tflog.Info(ctx, "fetching the incarnation", map[string]interface{}{"id": id})
		inc, clientErr = client.GetIncarnation(ctx, id)
		if errors.Is(clientErr, ErrNotFound) {
			err = ErrNotFound
			return
		}
		if clientErr != nil {
			diags.AddError("failed to retrieve incarnation", clientErr.Error())
		}
		return
	}

	timeout := 10 * time.Second
	if !waitForStatus.Timeout.IsNull() {
		timeout, clientErr = time.ParseDuration(waitForStatus.Timeout.ValueString())
		if clientErr != nil {
			diags.AddError("invalid timeout", clientErr.Error())
			return
		}
	}
	status := waitForStatus.Status.ValueString()
	tflog.Info(
		ctx,
		"fetching the incarnation",
		map[string]interface{}{
			"id":      id,
			"status":  status,
			"timeout": timeout.String(),
		},
	)

	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	inc, clientErr = client.GetIncarnationWithMergeRequestStatus(timeoutCtx, id, status)
	if inc.MergeRequestId == nil {
		tflog.Info(
			ctx,
			"No merge request in progress",
			map[string]interface{}{
				"id":      inc.Id,
				"details": "Since no merge request was initiated for the incarnation, it was not possible to wait for the requested status.",
			},
		)
	}
	if errors.Is(clientErr, ErrNotFound) {
		err = ErrNotFound
		return
	}
	if clientErr != nil {
		if errors.Is(clientErr, context.DeadlineExceeded) {
			diags.AddError("operation timed out before the merge request status reached the expected status", clientErr.Error())
			return
		}
		diags.AddError("failed to retrieve incarnation", clientErr.Error())
	}
	return
}
