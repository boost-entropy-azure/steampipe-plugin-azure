select name, akas, tags, title
from azure.azure_key_vault_certificate
where name = 'dummy-{{resourceName}}' and vault_name = '{{resourceName}}'
