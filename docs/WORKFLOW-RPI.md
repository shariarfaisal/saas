# Workflow: Research-Plan-Implement (RPI)

This is the recommended workflow for all development work on the Munchies SaaS platform.

## Phase 1: Research

Start by understanding the context and requirements.

### Steps

1. **Explore Requirements** - Read relevant docs in `docs/requirements/`
2. **Understand Current State** - Check existing implementations
3. **Identify Dependencies** - Look for related features or systems
4. **Gather Context** - Use Explore agent for large codebase searches
5. **Document Findings** - Note key constraints and patterns

### Example

If implementing a new API endpoint:

- Read `10-api-design.md` for standards
- Check `04-domain-model.md` for entities
- Review `03-multi-tenancy.md` for tenant handling
- Look at similar existing endpoints

## Phase 2: Plan

Design the solution before implementing.

### Steps

1. **Use EnterPlanMode** - Start formal planning for significant changes
2. **Consider Trade-offs** - Evaluate different approaches
3. **Identify Files to Change** - List all files that need modification
4. **Verify Architecture** - Ensure decisions align with system design
5. **Get Approval** - ExitPlanMode to request user review

### What to Include

- Architectural decisions and why
- Files that will be modified
- Breaking changes (if any)
- Testing strategy
- Documentation updates needed

## Phase 3: Implement

Execute the plan step by step.

### Steps

1. **Create TodoWrite List** - Break work into trackable steps
2. **Mark Tasks In Progress** - Update todos as work progresses
3. **Mark Tasks Complete** - Complete tasks immediately when done
4. **Commit Regularly** - Make focused commits with clear messages
5. **Test Thoroughly** - Verify changes work as expected

### Best Practices

- One task in progress at a time
- Never skip important exploration
- Always read files before editing
- Respect multi-tenancy requirements
- Reference requirements in commits

## Example Workflow

```
User: "Add notification system for order updates"
     ↓
Research: Read notifications.md, domain-model.md, check existing patterns
     ↓
Plan: Use EnterPlanMode to design service architecture, get approval
     ↓
Implement: Create todos, implement notification service step by step
     ↓
Complete: All tasks finished, tests passing, PR ready
```

---

Reference: best-practice from https://github.com/shanraisshan/claude-code-best-practice
