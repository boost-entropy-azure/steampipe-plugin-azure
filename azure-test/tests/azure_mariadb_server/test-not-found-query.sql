select
  name,
  id
from
  azure.azure_mariadb_server
where
  name = 'dummy-{{ resourceName }}'
  and resource_group = '{{ resourceName }}';