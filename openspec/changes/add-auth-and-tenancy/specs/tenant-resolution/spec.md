## ADDED Requirements

### Requirement: Tenant Resolution
The system SHALL resolve the active tenant on every inbound HTTP request before the handler is called.
Resolution SHALL follow this priority order: (1) subdomain from `Host` header, (2) `tenant_id` JWT
access-token claim, (3) `X-Tenant-ID` request header. The resolved tenant SHALL be cached in Redis for
60 seconds. Suspended or cancelled tenants SHALL receive a 403 response.

#### Scenario: Subdomain resolution success
- **WHEN** request arrives at `acme.api.platform.com/api/v1/...`
- **THEN** slug `acme` is extracted, tenant is looked up, and injected into context

#### Scenario: Suspended tenant blocked
- **WHEN** resolved tenant has status `suspended` or `cancelled`
- **THEN** middleware returns HTTP 403 with error code `TENANT_SUSPENDED`

#### Scenario: Unknown tenant
- **WHEN** slug/ID does not match any tenant in the database
- **THEN** middleware returns HTTP 404 with error code `TENANT_NOT_FOUND`
