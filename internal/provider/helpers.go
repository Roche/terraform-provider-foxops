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
	diags diag.Diagnostics,
	id IncarnationId,
	waitForStatus *waitForStatusMRModel,
) (inc Incarnation) {
	var err error
	if waitForStatus == nil {
		tflog.Info(ctx, "fetching the incarnation", map[string]interface{}{"id": id})
		inc, err = client.GetIncarnation(ctx, id)
		if err != nil {
			diags.AddError("failed to retrieve incarnation", err.Error())
			return
		}
	} else {
		timeout := 10 * time.Second
		if !waitForStatus.Timeout.IsNull() {
			timeout, err = time.ParseDuration(waitForStatus.Timeout.ValueString())
			if err != nil {
				diags.AddError("invalid timeout", err.Error())
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
		inc, err = client.GetIncarnationWithMergeRequestStatus(timeoutCtx, id, status)
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
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				diags.AddError("operation timed out before the merge request status reached the expected status", err.Error())
				return
			}
			diags.AddError("failed to retrieve incarnation", err.Error())
			return
		}
	}

	return
}
