# aistd project rules

Every commit must increase the affected application's SemVer VERSION and update its CHANGELOG.md; shared changes affect both CLI and website. Follow `docs/versioning.md`, preserve historical entries, and add a change fragment for user-facing changes. `make setup` installs the local pre-commit check; CI checks each new commit with the same script.

Read `apps/cli/AGENTS.md` before changing the generator. The ZIP files in `apps/cli/internal/standards/assets/` are pinned v1.0 source snapshots; a standard revision requires new hashes, tests, and explicit migration behavior. Keep the editable `task.md` current with milestones, task status, blockers, and verification evidence; it is not a pinned source.

For the website, use `docs/design/visual-spec.md` and `design/tokens/site.json`. Keep release and standards-status copy aligned with working behavior. Run `make check` and `make build` for code changes. `doctor` verifies installed standard sources, not project compliance.
