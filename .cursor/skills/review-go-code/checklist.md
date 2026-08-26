# Go bug patterns (gitboard)

Quick scan after reading changed code. Each item: pass / fail / N/A.

## Flow and errors

- [ ] All errors wrapped with context (`fmt.Errorf("...: %w", err)`)
- [ ] No `_ =` on errors that affect correctness
- [ ] No errors logged/printed without returning when caller must fail
- [ ] Unknown CLI commands exit 2 with usage; `main` owns fatal exits

## Resources

- [ ] Files closed (`defer f.Close()` after successful open)
- [ ] `scanner.Err()` checked after `bufio.Scanner` loops
- [ ] Context cancel deferred when created
- [ ] Goroutines have wait/cancel or clear ownership

## Forge and local git

- [ ] Process exec via `cliexec` / forge / localgit helpers (not scattered `exec.Command`)
- [ ] No live forge calls in unit tests (fakes or recorded output)
- [ ] Pull paths are ff-only; prune re-validates safety
- [ ] Worktree / dirty / upstream divergence handled without panic

## Variables and logic

- [ ] Variables created are used (no dead loads)
- [ ] No index out of range on slices
- [ ] Flag parsing covers `-h` / `--help` and unknown flags
- [ ] Named returns: no shadowing `err` when defer observes it

## Security

- [ ] No secrets or keys in source
- [ ] No committed real tokens or private forge payloads in fixtures
- [ ] Default listen remains loopback unless explicitly changed

## Tooling

- [ ] `make test` passes
- [ ] `make vet` passes
- [ ] `make ci` considered for surface changes
