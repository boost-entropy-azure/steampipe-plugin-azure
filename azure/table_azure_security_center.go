package azure

import (
	"context"
	"strings"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/resources/mgmt/policy"
	"github.com/Azure/azure-sdk-for-go/services/preview/security/mgmt/v1.0/security"
	"github.com/turbot/go-kit/types"
	"github.com/turbot/steampipe-plugin-sdk/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/plugin/transform"

	"github.com/turbot/steampipe-plugin-sdk/plugin"
)

//// TABLE DEFINITION

func tableAzureSecurityCenter(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "azure_security_center",
		Description: "Azure Security Center",
		List: &plugin.ListConfig{
			Hydrate: getSecurityCenterDetails,
		},
		Columns: []*plugin.Column{
			{
				Name:        "auto_provisioning_settings",
				Description: "Auto provisioning settings of the subscriptions.",
				Type:        proto.ColumnType_JSON,
				Hydrate:     getProvisioningDetails,
				Transform:   transform.FromValue(),
			},
			{
				Name:        "contacts",
				Description: "Security contact configurations for the subscription.",
				Type:        proto.ColumnType_JSON,
				Hydrate:     getContactDetails,
				Transform:   transform.FromValue(),
			},
			{
				Name:        "policy",
				Description: "Provides operations to assign policy definitions to a scope in your subscription.",
				Type:        proto.ColumnType_JSON,
				Hydrate:     getPolicyDetails,
				Transform:   transform.FromValue(),
			},
			{
				Name:        "pricings",
				Description: "Security pricing configuration in the resource group.",
				Type:        proto.ColumnType_JSON,
				Hydrate:     getPricingsDetails,
				Transform:   transform.FromValue(),
			},
			{
				Name:        "settings",
				Description: "Configuration settings for azure security center.",
				Type:        proto.ColumnType_JSON,
				Hydrate:     getSettingDetails,
				Transform:   transform.FromValue(),
			},

			// Steampipe standard columns
			{
				Name:        "title",
				Description: ColumnDescriptionTitle,
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromConstant("Security Center"),
			},
			{
				Name:        "akas",
				Description: ColumnDescriptionAkas,
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromValue().Transform(securityCenterAkas),
			},

			// Azure standard columns
			{
				Name:        "subscription_id",
				Description: ColumnDescriptionSubscription,
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromValue(),
			},
		},
	}
}

//// LIST FUNCTION

func getSecurityCenterDetails(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	session, err := GetNewSession(ctx, d, "MANAGEMENT")
	if err != nil {
		return nil, err
	}

	subscriptionID := session.SubscriptionID
	d.StreamListItem(ctx, subscriptionID)

	return nil, nil
}

//// HYDRATE FUNCTIONS

func getProvisioningDetails(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	session, err := GetNewSession(ctx, d, "MANAGEMENT")
	if err != nil {
		return nil, err
	}

	subscriptionID := session.SubscriptionID
	autoProvisionClient := security.NewAutoProvisioningSettingsClient(subscriptionID, "")
	autoProvisionClient.Authorizer = session.Authorizer

	autoProvisionList, err := autoProvisionClient.List(ctx)
	if err != nil {
		return err, nil
	}

	// If we return the API response directly, the output only gives the contents of AutoProvisionList
	var provisioning []map[string]interface{}

	for _, provision := range autoProvisionList.Values() {
		provisionMap := make(map[string]interface{})
		provisionMap["id"] = provision.ID
		provisionMap["name"] = provision.Name
		provisionMap["properties"] = provision.AutoProvisioningSettingProperties
		provisionMap["type"] = provision.Type
		provisioning = append(provisioning, provisionMap)
	}
	return provisioning, nil
}

func getContactDetails(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	session, err := GetNewSession(ctx, d, "MANAGEMENT")
	if err != nil {
		return nil, err
	}

	subscriptionID := session.SubscriptionID
	contactClient := security.NewContactsClient(subscriptionID, "")
	contactClient.Authorizer = session.Authorizer

	contactList, err := contactClient.List(ctx)
	if err != nil {
		return err, nil
	}

	// If we return the API response directly, the output only gives the contents of contactList
	var contacts []map[string]interface{}

	for _, contact := range contactList.Values() {
		contactMap := make(map[string]interface{})
		contactMap["id"] = contact.ID
		contactMap["name"] = contact.Name
		contactMap["properties"] = contact.ContactProperties
		contactMap["type"] = contact.Type
		contacts = append(contacts, contactMap)
	}
	return contacts, nil
}

func getPolicyDetails(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	session, err := GetNewSession(ctx, d, "MANAGEMENT")
	if err != nil {
		return nil, err
	}

	subscriptionID := session.SubscriptionID
	PolicyClient := policy.NewAssignmentsClient(subscriptionID)
	PolicyClient.Authorizer = session.Authorizer

	policy, err := PolicyClient.Get(ctx, "/subscriptions/"+subscriptionID, "SecurityCenterBuiltIn")
	if err != nil {
		return err, nil
	}

	return policy, nil
}

func getSettingDetails(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	session, err := GetNewSession(ctx, d, "MANAGEMENT")
	if err != nil {
		return nil, err
	}
	subscriptionID := session.SubscriptionID
	settingClient := security.NewSettingsClient(subscriptionID, "")
	settingClient.Authorizer = session.Authorizer

	settingList, err := settingClient.List(ctx)
	if err != nil {
		return err, nil
	}

	// If we return the API response directly, the output only gives the contents of SettingsList
	var settings []map[string]interface{}

	for _, setting := range settingList.Values() {
		settingMap := make(map[string]interface{})
		settingMap["id"] = setting.ID
		settingMap["name"] = setting.Name
		settingMap["kind"] = setting.Kind
		settingMap["type"] = setting.Type
		settings = append(settings, settingMap)
	}
	return settings, nil
}

func getPricingsDetails(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	session, err := GetNewSession(ctx, d, "MANAGEMENT")
	if err != nil {
		return nil, err
	}

	subscriptionID := session.SubscriptionID
	pricingClient := security.NewPricingsClient(subscriptionID, "")
	pricingClient.Authorizer = session.Authorizer

	pricingList, err := pricingClient.List(ctx)
	if err != nil {
		return err, nil
	}

	// If we return the API response directly, the output only gives the contents of PricingList
	var pricings []map[string]interface{}

	for _, pricing := range pricingList.Values() {
		pricingMap := make(map[string]interface{})
		pricingMap["id"] = pricing.ID
		pricingMap["name"] = pricing.Name
		pricingMap["properties"] = pricing.PricingProperties
		pricingMap["type"] = pricing.Type
		pricings = append(pricings, pricingMap)
	}
	return pricings, nil
}

//// TRANSFORM FUNCTIONS

func securityCenterAkas(ctx context.Context, d *transform.TransformData) (interface{}, error) {
	subscriptionID := types.SafeString(d.Value)
	id := "/subscriptions/" + subscriptionID + "/providers/Microsoft.Security/securityCenter"
	akas := []string{"azure://" + id, "azure://" + strings.ToLower(id)}
	return akas, nil
}