# Code Organization Rule

All code should follow the project's established patterns:

- Each subsystem (api, website, portal, admin) has its own repo
- Shared concerns are documented in requirements
- Multi-tenancy is enforced at API and database layers
- Configuration is environment-based, not code-based

## File Paths

When working in specific subsystems:

- **API work**: Reference `munchies-api` folder structure
- **Web work**: Reference `Munchies-website-app` folder structure
- **Portal work**: Reference `restaurant-portal` folder structure
- **Admin work**: Reference `madmin` folder structure

## Cross-Subsystem Coordination

When changes span multiple systems:

- Create OpenSpec proposal if breaking changes
- Document API contracts in API design spec
- Sync database migrations across systems
- Update shared documentation

---

Path scope: All subsystems
