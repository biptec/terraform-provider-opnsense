package acmeclient

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	apiacme "github.com/biptec/opnsense-go/pkg/acmeclient"
	"github.com/biptec/opnsense-go/pkg/api"
	"github.com/biptec/opnsense-go/pkg/errs"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func boolString(v bool) string {
	if v {
		return "1"
	}
	return "0"
}
func stringBool(v string) bool { return v == "1" || strings.EqualFold(v, "true") }

func setStrings(ctx context.Context, value types.Set) ([]string, error) {
	if value.IsNull() || value.IsUnknown() {
		return nil, nil
	}
	var result []string
	diags := value.ElementsAs(ctx, &result, false)
	if diags.HasError() {
		return nil, fmt.Errorf("unable to decode string set: %s", diags.Errors()[0].Detail())
	}
	return result, nil
}

func stringSet(values []string) types.Set {
	result, _ := types.SetValueFrom(context.Background(), types.StringType, values)
	return result
}

func terminalACMEStatus(status string) bool {
	return status == "200" || status == "250" || status == "300" || status == "400" || status == "500"
}

func waitAccountRegistration(ctx context.Context, client *apiacme.Controller, id string) (*apiacme.Account, error) {
	deadline := time.NewTimer(90 * time.Second)
	defer deadline.Stop()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		current, err := client.GetAccount(ctx, id)
		if err != nil {
			return nil, err
		}
		if current.StatusCode == "200" {
			return current, nil
		}
		if terminalACMEStatus(current.StatusCode) && current.StatusCode != "100" && current.StatusCode != "" {
			return current, fmt.Errorf("ACME account registration ended with status code %s", current.StatusCode)
		}
		select {
		case <-ctx.Done():
			return current, ctx.Err()
		case <-deadline.C:
			return current, fmt.Errorf("timed out waiting for ACME account registration; last status=%s", current.StatusCode)
		case <-ticker.C:
		}
	}
}

func waitCertificateIssued(ctx context.Context, client *apiacme.Controller, id string) (*apiacme.Certificate, error) {
	deadline := time.NewTimer(5 * time.Minute)
	defer deadline.Stop()
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		current, err := client.GetCertificate(ctx, id)
		if err != nil {
			return nil, err
		}
		if current.StatusCode == "200" && current.CertRefID != "" {
			return current, nil
		}
		if terminalACMEStatus(current.StatusCode) && current.StatusCode != "100" && current.StatusCode != "" {
			return current, fmt.Errorf("ACME certificate issuance ended with status code %s", current.StatusCode)
		}
		select {
		case <-ctx.Done():
			return current, ctx.Err()
		case <-deadline.C:
			return current, fmt.Errorf("timed out waiting for ACME certificate issuance; last status=%s", current.StatusCode)
		case <-ticker.C:
		}
	}
}

func accountToAPI(d *accountResourceModel) *apiacme.Account {
	return &apiacme.Account{
		Enabled: boolString(d.Enabled.ValueBool()), Name: d.Name.ValueString(), Description: d.Description.ValueString(), Email: d.Email.ValueString(), CA: api.SelectedMap(d.CA.ValueString()),
	}
}
func accountFromAPI(id string, d *apiacme.Account, prior *accountResourceModel) *accountResourceModel {
	registered := d.StatusCode == "200"
	return &accountResourceModel{ID: types.StringValue(id), Enabled: types.BoolValue(stringBool(d.Enabled)), Name: types.StringValue(d.Name), Description: types.StringValue(d.Description), Email: types.StringValue(d.Email), CA: types.StringValue(d.CA.String()), Register: prior.Register, RegistrationVersion: prior.RegistrationVersion, Registered: types.BoolValue(registered), StatusCode: types.StringValue(d.StatusCode), StatusLastUpdate: types.StringValue(d.StatusLastUpdate)}
}

func validationToAPI(d *validationResourceModel, key string) *apiacme.Validation {
	return &apiacme.Validation{
		Enabled: boolString(d.Enabled.ValueBool()), Name: d.Name.ValueString(), Description: d.Description.ValueString(), Method: api.SelectedMap(d.Method.ValueString()), DNSService: api.SelectedMap(d.DNSService.ValueString()), DNSNsupdateServer: d.DNSNsupdateServer.ValueString(), DNSNsupdateZone: d.DNSNsupdateZone.ValueString(), DNSNsupdateKey: key,
	}
}
func validationFromAPI(id string, d *apiacme.Validation, prior *validationResourceModel) *validationResourceModel {
	return &validationResourceModel{
		ID: types.StringValue(id), Enabled: types.BoolValue(stringBool(d.Enabled)), Name: types.StringValue(d.Name), Description: types.StringValue(d.Description), Method: types.StringValue(d.Method.String()), DNSService: types.StringValue(d.DNSService.String()), DNSNsupdateServer: types.StringValue(d.DNSNsupdateServer), DNSNsupdateZone: types.StringValue(d.DNSNsupdateZone), DNSNsupdateKey: types.StringNull(), DNSNsupdateKeyVersion: prior.DNSNsupdateKeyVersion, DNSNsupdateKeyConfigured: types.BoolValue(d.DNSNsupdateKey != ""),
	}
}

func actionToAPI(d *actionResourceModel) *apiacme.Action {
	return &apiacme.Action{Enabled: boolString(d.Enabled.ValueBool()), Name: d.Name.ValueString(), Description: d.Description.ValueString(), Type: api.SelectedMap(d.Type.ValueString())}
}
func actionFromAPI(id string, d *apiacme.Action) *actionResourceModel {
	return &actionResourceModel{ID: types.StringValue(id), Enabled: types.BoolValue(stringBool(d.Enabled)), Name: types.StringValue(d.Name), Description: types.StringValue(d.Description), Type: types.StringValue(d.Type.String())}
}

func certificateToAPI(ctx context.Context, d *certificateResourceModel) (*apiacme.Certificate, error) {
	alt, err := setStrings(ctx, d.AltNames)
	if err != nil {
		return nil, err
	}
	actions, err := setStrings(ctx, d.RestartActionIDs)
	if err != nil {
		return nil, err
	}
	return &apiacme.Certificate{Enabled: boolString(d.Enabled.ValueBool()), Name: d.Name.ValueString(), Description: d.Description.ValueString(), AltNames: api.SelectedMapList(alt), Account: api.SelectedMap(d.AccountID.ValueString()), ValidationMethod: api.SelectedMap(d.ValidationID.ValueString()), KeyLength: api.SelectedMap(d.KeyLength.ValueString()), OCSP: "0", RestartActions: api.SelectedMapList(actions), AutoRenewal: boolString(d.AutoRenewal.ValueBool()), RenewInterval: strconv.FormatInt(d.RenewInterval.ValueInt64(), 10), AliasMode: api.SelectedMap(d.AliasMode.ValueString()), ChallengeAlias: d.ChallengeAlias.ValueString()}, nil
}
func certificateFromAPI(id string, d *apiacme.Certificate, prior *certificateResourceModel) *certificateResourceModel {
	interval, _ := strconv.ParseInt(d.RenewInterval, 10, 64)
	if interval < 0 {
		interval = 0
	}
	return &certificateResourceModel{ID: types.StringValue(id), Enabled: types.BoolValue(stringBool(d.Enabled)), Name: types.StringValue(d.Name), Description: types.StringValue(d.Description), AltNames: stringSet([]string(d.AltNames)), AccountID: types.StringValue(d.Account.String()), ValidationID: types.StringValue(d.ValidationMethod.String()), KeyLength: types.StringValue(d.KeyLength.String()), RestartActionIDs: stringSet([]string(d.RestartActions)), AutoRenewal: types.BoolValue(stringBool(d.AutoRenewal)), RenewInterval: types.Int64Value(interval), AliasMode: types.StringValue(d.AliasMode.String()), ChallengeAlias: types.StringValue(d.ChallengeAlias), Issue: prior.Issue, IssuanceVersion: prior.IssuanceVersion, CertRefID: types.StringValue(d.CertRefID), Issued: types.BoolValue(d.StatusCode == "200" && d.CertRefID != ""), LastUpdate: types.StringValue(d.LastUpdate), StatusCode: types.StringValue(d.StatusCode), StatusLastUpdate: types.StringValue(d.StatusLastUpdate)}
}

func isNotFound(err error) bool { var nf *errs.NotFoundError; return errors.As(err, &nf) }
