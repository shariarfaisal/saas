# Munchies SaaS - Claude Code Configuration

This is the persistent context file for AI assistants working on the Munchies SaaS platform. Keep this file under 150 lines for optimal performance.

## Project Overview

Munchies is a multi-tenant SaaS platform for restaurant management. The project is divided into:

- **munchies-api** - Backend API (Node.js)
- **Munchies-website-app** - Frontend Web App
- **restaurant-portal** - Restaurant Management Portal
- **madmin** - Admin Dashboard

This folder contains the overall architecture, specifications, and documentation.

## OpenSpec System

Always open `@/openspec/AGENTS.md` when the request:

- Mentions planning or proposals (words like proposal, spec, change, plan)
- Introduces new capabilities, breaking changes, architecture shifts, or big performance/security work
- Sounds ambiguous and you need the authoritative spec before coding

Use openspec to learn about creating and applying change proposals with proper spec format.

## Key Documentation

- **docs/requirements/** - Complete requirements documentation for all platform aspects
- **openspec/project.md** - Project structure and guidelines
- **.github/prompts/** - OpenSpec proposal, archive, and apply prompts

## Workflow Patterns (RPI Methodology)

1. **Research** - Explore codebase, understand context, read requirements
2. **Plan** - Use EnterPlanMode to design approach before coding
3. **Implement** - Execute the plan step by step, marking todos

## Best Practices

- Use Task tool with Explore agent for large codebase searches
- Break complex tasks into smaller steps using TodoWrite
- Always read files before modifying them
- Use AskUserQuestion to clarify requirements
- Keep git commits focused and meaningful
- Never skip exploration in requirements documentation

## Debugging & Tools

- Use `/doctor` to check Claude Code configuration
- Use background tasks for long-running operations
- Use MCP servers: Context7, Playwright, Claude in Chrome, DeepWiki
- Use git commands for repository operations

## Multi-Tenant Architecture Notes

- The system supports multiple restaurant tenants
- Documentation covers: multi-tenancy (03), database schema (09), API design (10)
- Always verify tenant isolation requirements when implementing features
- Pricing and financials (08) are critical for SaaS features

---

Last updated: 2026-02-27
