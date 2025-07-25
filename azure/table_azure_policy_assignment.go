package azure

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/resources/mgmt/policy"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

//// TABLE DEFINITION

func tableAzurePolicyAssignment(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "azure_policy_assignment",
		Description: "Azure Policy Assignment",
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("id"),
			Hydrate:    getPolicyAssignment,
			Tags: map[string]string{
				"service": "Microsoft.Authorization",
				"action":  "policyAssignments/read",
			},
			IgnoreConfig: &plugin.IgnoreConfig{
				ShouldIgnoreErrorFunc: isNotFoundError([]string{"ResourceNotFound", "ResourceGroupNotFound"}),
			},
		},
		List: &plugin.ListConfig{
			Hydrate: listPolicyAssignments,
			Tags: map[string]string{
				"service": "Microsoft.Authorization",
				"action":  "policyAssignments/read",
			},
		},
		Columns: azureColumns([]*plugin.Column{
			{
				Name:        "id",
				Type:        proto.ColumnType_STRING,
				Description: "The ID of the policy assignment.",
				Transform:   transform.FromGo(),
			},
			{
				Name:        "name",
				Description: "The name of the policy assignment.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "display_name",
				Description: "The display name of the policy assignment.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("AssignmentProperties.DisplayName"),
			},
			{
				Name:        "policy_definition_id",
				Description: "The ID of the policy definition or policy set definition being assigned.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("AssignmentProperties.PolicyDefinitionID"),
			},
			{
				Name:        "description",
				Description: "This message will be part of response in case of policy violation.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("AssignmentProperties.Description"),
			},
			{
				Name:        "enforcement_mode",
				Description: "The policy assignment enforcement mode. Possible values are Default and DoNotEnforce.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("AssignmentProperties.EnforcementMode"),
			},
			{
				Name:        "scope",
				Description: "The scope for the policy assignment.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("AssignmentProperties.Scope"),
			},
			{
				Name:        "sku_name",
				Description: "The name of the policy sku.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Sku.Name"),
			},
			{
				Name:        "sku_tier",
				Description: "The policy sku tier.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Sku.Tier").Transform(transform.ToString),
			},
			{
				Name:        "type",
				Description: "The type of the policy assignment.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "identity",
				Description: "The managed identity associated with the policy assignment.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "metadata",
				Description: "The policy assignment metadata.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("AssignmentProperties.Metadata"),
			},
			{
				Name:        "not_scopes",
				Description: "The policy's excluded scopes.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("AssignmentProperties.NotScopes"),
			},
			{
				Name:        "parameters",
				Description: "The parameter values for the assigned policy rule.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("AssignmentProperties.Parameters"),
			},

			// Steampipe standard columns
			{
				Name:        "title",
				Description: ColumnDescriptionTitle,
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Name"),
			},
			{
				Name:        "akas",
				Description: ColumnDescriptionAkas,
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("ID").Transform(idToAkas),
			},
		}),
	}
}

//// LIST FUNCTION

func listPolicyAssignments(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	session, err := GetNewSession(ctx, d, "MANAGEMENT")
	if err != nil {
		return nil, err
	}

	subscriptionID := session.SubscriptionID
	policyClient := policy.NewAssignmentsClientWithBaseURI(session.ResourceManagerEndpoint, subscriptionID)
	policyClient.Authorizer = session.Authorizer

	// Apply Retry rule
	ApplyRetryRules(ctx, &policyClient, d.Connection)

	result, err := policyClient.List(ctx, "")
	if err != nil {
		return err, nil
	}

	for _, policy := range result.Values() {
		d.StreamListItem(ctx, policy)
		// Check if context has been cancelled or if the limit has been hit (if specified)
		// if there is a limit, it will return the number of rows required to reach this limit
		if d.RowsRemaining(ctx) == 0 {
			return nil, nil
		}
	}

	for result.NotDone() {
		// Wait for rate limiting
		d.WaitForListRateLimit(ctx)

		err = result.NextWithContext(ctx)
		if err != nil {
			return nil, err
		}

		for _, policy := range result.Values() {
			d.StreamListItem(ctx, policy)
			// Check if context has been cancelled or if the limit has been hit (if specified)
			// if there is a limit, it will return the number of rows required to reach this limit
			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}
	}

	return nil, nil
}

//// HYDRATE FUNCTIONS

func getPolicyAssignment(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	session, err := GetNewSession(ctx, d, "MANAGEMENT")
	if err != nil {
		return nil, err
	}
	id := d.EqualsQuals["id"].GetStringValue()

	subscriptionID := session.SubscriptionID
	policyClient := policy.NewAssignmentsClientWithBaseURI(session.ResourceManagerEndpoint, subscriptionID)
	policyClient.Authorizer = session.Authorizer

	// Apply Retry rule
	ApplyRetryRules(ctx, &policyClient, d.Connection)

	policy, err := policyClient.GetByID(ctx, id)
	if err != nil {
		return err, nil
	}

	return policy, nil
}
