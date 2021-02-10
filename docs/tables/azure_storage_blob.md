# Table: azure_storage_blob

Azure Blob Storage helps you create data lakes for your analytics needs and provides storage to build powerful cloud-native and mobile apps. Optimise costs with tiered storage for your long-term data and flexibly scale up for high-performance computing and machine learning workloads.

Note that you **_must_** specify a single `storage_account_name` and `resource_group` in a where clause in order to use this table. Also, see the note below on issue relating to a [known issue](https://github.com/turbot/steampipe-postgres-fdw/issues/3) with nested select queries (select where in (select ...)) and joins on tables with required key columns.

## Examples

### Get basic info for all the blobs

```sql
select
  name,
  storage_account_name,
  container_name,
  type,
  is_snapshot,
  access_tier
from
  azure_storage_blob
where
  storage_account_name = 'dummy'
  and resource_group = 'resource_test'
```

### List all the blobs of the type snapshot with import data

```sql
select
  name,
  type,
  access_tier,
  server_encrypted,
  metadata,
  creation_time,
  container_name,
  storage_account_name,
  resource_group,
  region
from
  azure_storage_blob
where
  storage_account_name = 'dummy'
  and resource_group = 'resource_test'
  and is_snapshot;
```

select
  name,
  type,
  access_tier,
  server_encrypted,
  metadata,
  creation_time,
  container_name,
  storage_account_name,
  resource_group,
  region
from
  azure_storage_blob
where
  storage_account_name = 'dummy'
  and resource_group = 'resource_test'
  and is_snapshot;
