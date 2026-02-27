# Munchies SaaS - Claude Code Best Practices

This file documents the best practices and patterns for AI-assisted development on the Munchies SaaS platform using Claude Code.

## Core Principles

1. **Always Research First** - Understand requirements and context before planning
2. **Plan Before Code** - Use EnterPlanMode for significant changes
3. **Respect Multi-Tenancy** - Every feature must handle multiple restaurants
4. **Use OpenSpec** - Create proposals for breaking changes and major features
5. **Document Decisions** - Reference requirements in code and commits

## Quick Reference

- **CLAUDE.md** - Persistent context and key information
- **docs/WORKFLOW-RPI.md** - Research-Plan-Implement workflow details
- **openspec/AGENTS.md** - How to create change proposals
- **docs/requirements/** - Complete feature and architectural documentation

## Tools & Agents

### Specialized Agents

- **spec-agent** - For architectural proposals and OpenSpec work
- **multi-tenancy-expert** - For verifying tenant isolation

### Skills (Reusable Knowledge)

- **api-design** - REST API standards and patterns
- **requirements-reference** - Complete documentation index

### Rules (Topic-Specific Guidelines)

- **code-organization** - File structure and subsystem coordination
- **database-queries** - Multi-tenancy and query standards
- **security-permissions** - Authorization and access control

## Common Workflows

### Adding a New Feature

1. Check `docs/requirements/05-feature-requirements.md`
2. Use EnterPlanMode to design approach
3. Reference relevant requirements throughout implementation
4. Create OpenSpec proposal if significant changes
5. Verify multi-tenancy compliance

### Fixing a Bug

1. Research the issue (use Explore agent if needed)
2. Identify root cause
3. Write test that reproduces the bug
4. Fix the bug
5. Verify test passes

### Refactoring

1. Understand current implementation
2. Use EnterPlanMode for architectural changes
3. Keep backward compatibility unless using OpenSpec
4. Update documentation if patterns change

## Key Documentation by Module

| Module         | Key Docs   |
| -------------- | ---------- |
| Database       | 09, 03, 04 |
| API            | 10, 04, 06 |
| Multi-Tenant   | 03, 06, 07 |
| Payments       | 08, 11     |
| Features       | 05, 14     |
| Infrastructure | 13, 15     |
| Analytics      | 12         |

## Getting Started

1. Read `CLAUDE.md` for project overview
2. Review `docs/WORKFLOW-RPI.md` for workflow patterns
3. Check relevant requirements for your feature
4. Use Task tool with Explore agent for codebase search
5. Use EnterPlanMode before significant changes

---

See also: `.claude/` directory for agents, skills, and rules
