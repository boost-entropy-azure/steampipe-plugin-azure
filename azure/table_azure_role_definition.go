package azure

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/authorization/mgmt/authorization"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"

	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
)

//// TABLE DEFINITION

func tableAzureRoleDefinition(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "azure_role_definition",
		Description: "Azure Role Definition",
		Get: &plugin.GetConfig{
			KeyColumns: plugin.SingleColumn("name"),
			Hydrate:    getIamRoleDefinition,
			Tags: map[string]string{
				"service": "Microsoft.Authorization",
				"action":  "roleDefinitions/read",
			},
			IgnoreConfig: &plugin.IgnoreConfig{
				ShouldIgnoreErrorFunc: isNotFoundError([]string{"ResourceNotFound", "ResourceGroupNotFound"}),
			},
		},
		List: &plugin.ListConfig{
			Hydrate: listIamRoleDefinitions,
			Tags: map[string]string{
				"service": "Microsoft.Authorization",
				"action":  "roleDefinitions/read",
			},
		},
		Columns: azureColumns([]*plugin.Column{
			{
				Name:        "name",
				Type:        proto.ColumnType_STRING,
				Description: "The friendly name that identifies the role definition.",
			},
			{
				Name:        "id",
				Description: "Contains ID to identify a role definition uniquely.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromGo(),
			},
			{
				Name:        "type",
				Description: "Contains the resource type.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "role_name",
				Description: "Current state of the role definition.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("RoleDefinitionProperties.RoleName"),
			},
			{
				Name:        "role_type",
				Description: "Name of the role definition.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("RoleDefinitionProperties.RoleType"),
			},
			{
				Name:        "description",
				Description: "Description of the role definition.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("RoleDefinitionProperties.Description"),
			},
			{
				Name:        "assignable_scopes",
				Description: "A list of assignable scopes for which the role definition can be assigned.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("RoleDefinitionProperties.AssignableScopes"),
			},
			{
				Name:        "permissions",
				Description: "A list of actions, which can be accessed.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("RoleDefinitionProperties.Permissions"),
			},
			{
				Name:        "title",
				Description: ColumnDescriptionTitle,
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("RoleDefinitionProperties.RoleName"),
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

func listIamRoleDefinitions(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	session, err := GetNewSession(ctx, d, "MANAGEMENT")
	if err != nil {
		return nil, err
	}
	subscriptionID := session.SubscriptionID
	authorizationClient := authorization.NewRoleDefinitionsClientWithBaseURI(session.ResourceManagerEndpoint, subscriptionID)
	authorizationClient.Authorizer = session.Authorizer

	// Apply Retry rule
	ApplyRetryRules(ctx, &authorizationClient, d.Connection)

	result, err := authorizationClient.List(ctx, "/subscriptions/"+subscriptionID, "")
	if err != nil {
		return nil, err
	}
	for _, roleDefinition := range result.Values() {
		d.StreamListItem(ctx, roleDefinition)
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

		for _, roleDefinition := range result.Values() {
			d.StreamListItem(ctx, roleDefinition)
			// Check if context has been cancelled or if the limit has been hit (if specified)
			// if there is a limit, it will return the number of rows required to reach this limit
			if d.RowsRemaining(ctx) == 0 {
				return nil, nil
			}
		}
	}

	return nil, err
}

//// HYDRATE FUNCTIONS

func getIamRoleDefinition(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	plugin.Logger(ctx).Trace("getIamRoleDefinition")

	session, err := GetNewSession(ctx, d, "MANAGEMENT")
	if err != nil {
		return nil, err
	}
	subscriptionID := session.SubscriptionID
	name := d.EqualsQuals["name"].GetStringValue()

	authorizationClient := authorization.NewRoleDefinitionsClientWithBaseURI(session.ResourceManagerEndpoint, subscriptionID)
	authorizationClient.Authorizer = session.Authorizer

	// Apply Retry rule
	ApplyRetryRules(ctx, &authorizationClient, d.Connection)

	op, err := authorizationClient.Get(ctx, "/subscriptions/"+subscriptionID, name)
	if err != nil {
		return nil, err
	}

	return op, nil
}
