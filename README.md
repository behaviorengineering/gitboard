# Gitboard

Local code-change board for GitLab and GitHub: CI status, open MRs/PRs, failed job triage. Uses `gh` and `glab` on your machine (not deployments or runtime ops).

**Remote:** [github.com/behaviorengineering/gitboard](https://github.com/behaviorengineering/gitboard)  
**Consilium path (in-tree copy):** `providers/gitboard/` until submodule is wired.

## Quick start

```bash
cp projects.example.yaml projects.yaml
brew install glab
glab auth login
gh auth login
make build serve
```

Open [http://127.0.0.1:1325/](http://127.0.0.1:1325/).

## Design

| Layer | Choice |
|-------|--------|
| GitHub | [`gh`](https://cli.github.com/) |
| GitLab | [`glab`](https://gitlab.com/gitlab-org/cli) |
| Auth | Existing CLI login sessions |
| AI triage | Optional (`GITBOARD_LLM_BASE_URL` or `POLYPUS_BASE_URL`) |

## Projects file

`projects.yaml` (gitignored) lists tracked repos. See `projects.example.yaml`.

Override with `GITBOARD_PROJECTS`.

## API

| Method | Path | Purpose |
|--------|------|---------|
| `GET` | `/api/health` | Liveness |
| `GET` | `/api/dashboard` | Aggregated rows |
| `GET` | `/api/failures?project=&run_id=` | Failed jobs |
| `POST` | `/api/triage` | AI log analysis |

## Related

| Piece | Location |
|-------|----------|
| Consilium docs | [docs/run/providers/gitboard/](../../docs/run/providers/gitboard/) (product git) |
