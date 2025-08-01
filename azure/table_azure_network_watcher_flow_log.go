package azure

import (
	"context"
	"strings"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/network/mgmt/network"
	"github.com/turbot/steampipe-plugin-sdk/v5/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/transform"
)

type flowLogInfo = struct {
	network.FlowLog
	NetworkWatcherName string
}

//// TABLE DEFINITION

func tableAzureNetworkWatcherFlowLog(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "azure_network_watcher_flow_log",
		Description: "Azure Network Watcher Flow Log",
		Get: &plugin.GetConfig{
			KeyColumns: plugin.AllColumns([]string{"network_watcher_name", "name", "resource_group"}),
			Hydrate:    getNetworkWatcherFlowLog,
			Tags: map[string]string{
				"service": "Microsoft.Network",
				"action":  "networkWatchers/flowLogs/read",
			},
			IgnoreConfig: &plugin.IgnoreConfig{
				ShouldIgnoreErrorFunc: isNotFoundError([]string{"ResourceNotFound", "ResourceGroupNotFound", "404"}),
			},
		},
		List: &plugin.ListConfig{
			ParentHydrate: listNetworkWatchers,
			Hydrate: listNetworkWatcherFlowLogs,
			Tags: map[string]string{
				"service": "Microsoft.Network",
				"action":  "networkWatchers/flowLogs/read",
			},
		},
		Columns: azureColumns([]*plugin.Column{
			{
				Name:        "name",
				Description: "The friendly name that identifies the flow log.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "id",
				Description: "Contains ID to identify a flow log uniquely.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromGo(),
			},
			{
				Name:        "enabled",
				Description: "Indicates whether the flow log is enabled, or not.",
				Type:        proto.ColumnType_BOOL,
				Transform:   transform.FromField("FlowLogPropertiesFormat.Enabled"),
			},
			{
				Name:        "network_watcher_name",
				Description: "The friendly name that identifies the network watcher.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "provisioning_state",
				Description: "The provisioning state of the flow log.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("FlowLogPropertiesFormat.ProvisioningState").Transform(transform.ToString),
			},
			{
				Name:        "type",
				Description: "The resource type of the flow log.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "version",
				Description: "The version (revision) of the flow log.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("FlowLogPropertiesFormat.Format.Version"),
			},
			{
				Name:        "etag",
				Description: "An unique read-only string that changes whenever the resource is updated.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "file_type",
				Description: "The file type of flow log. Possible values include: 'JSON'.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("FlowLogPropertiesFormat.Format.Type").Transform(transform.ToString),
			},
			{
				Name:        "retention_policy_days",
				Description: "Specifies the number of days to retain flow log records.",
				Type:        proto.ColumnType_INT,
				Transform:   transform.FromField("FlowLogPropertiesFormat.RetentionPolicy.Days"),
			},
			{
				Name:        "retention_policy_enabled",
				Description: "Indicates whether flow log retention is enabled, or not.",
				Type:        proto.ColumnType_BOOL,
				Transform:   transform.FromField("FlowLogPropertiesFormat.RetentionPolicy.Enabled"),
			},
			{
				Name:        "storage_id",
				Description: "The ID of the storage account which is used to store the flow log.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("FlowLogPropertiesFormat.StorageID"),
			},
			{
				Name:        "target_resource_id",
				Description: "The ID of network security group to which flow log will be applied.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("FlowLogPropertiesFormat.TargetResourceID"),
			},
			{
				Name:        "target_resource_guid",
				Description: "The Guid of network security group to which flow log will be applied.",
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("FlowLogPropertiesFormat.TargetResourceGUID"),
			},
			{
				Name:        "traffic_analytics",
				Description: "Defines the configuration of flow log traffic analytics.",
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("FlowLogPropertiesFormat.FlowAnalyticsConfiguration.NetworkWatcherFlowAnalyticsConfiguration"),
			},

			// Steampipe standard columns
			{
				Name:        "title",
				Description: ColumnDescriptionTitle,
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Name"),
			},
			{
				Name:        "tags",
				Description: ColumnDescriptionTags,
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "akas",
				Description: ColumnDescriptionAkas,
				Type:        proto.ColumnType_JSON,
				Transform:   transform.FromField("ID").Transform(idToAkas),
			},

			// Azure standard columns
			{
				Name:        "region",
				Description: ColumnDescriptionRegion,
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("Location").Transform(toLower),
			},
			{
				Name:        "resource_group",
				Description: ColumnDescriptionResourceGroup,
				Type:        proto.ColumnType_STRING,
				Transform:   transform.FromField("ID").Transform(extractResourceGroupFromID),
			},
		}),
	}
}

//// LIST FUNCTIONS

func listNetworkWatcherFlowLogs(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	// Get the details of network watcher
	networkWatcherDetails := h.Item.(network.Watcher)

	// Create session
	session, err := GetNewSession(ctx, d, "MANAGEMENT")
	if err != nil {
		return nil, err
	}
	subscriptionID := session.SubscriptionID
	resourceGroupID := strings.Split(*networkWatcherDetails.ID, "/")[4]

	client := network.NewFlowLogsClientWithBaseURI(session.ResourceManagerEndpoint, subscriptionID)
	client.Authorizer = session.Authorizer

	// Apply Retry rule
	ApplyRetryRules(ctx, &client, d.Connection)

	result, err := client.List(ctx, resourceGroupID, *networkWatcherDetails.Name)
	if err != nil {
		return nil, err
	}
	for _, flowLog := range result.Values() {
		d.StreamListItem(ctx, flowLogInfo{flowLog, *networkWatcherDetails.Name})
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

		for _, flowLog := range result.Values() {
			d.StreamListItem(ctx, flowLogInfo{flowLog, *networkWatcherDetails.Name})
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

func getNetworkWatcherFlowLog(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	plugin.Logger(ctx).Trace("getNetworkWatcherFlowLog")

	networkWatcherName := d.EqualsQuals["network_watcher_name"].GetStringValue()
	name := d.EqualsQuals["name"].GetStringValue()
	resourceGroup := d.EqualsQuals["resource_group"].GetStringValue()

	session, err := GetNewSession(ctx, d, "MANAGEMENT")
	if err != nil {
		return nil, err
	}
	subscriptionID := session.SubscriptionID

	client := network.NewFlowLogsClientWithBaseURI(session.ResourceManagerEndpoint, subscriptionID)
	client.Authorizer = session.Authorizer

	// Apply Retry rule
	ApplyRetryRules(ctx, &client, d.Connection)

	op, err := client.Get(ctx, resourceGroup, networkWatcherName, name)
	if err != nil {
		return nil, err
	}

	return flowLogInfo{op, networkWatcherName}, nil
}
