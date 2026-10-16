## Description
Briefly describe the change, rationale, and impact.

## Type of Change
- [ ] Bug fix (`fix:`)
- [ ] New feature (`feat:`)
- [ ] Documentation update (`docs:`)
- [ ] Testing improvement (`test:`)
- [ ] Refactor / Performance (`refactor:`)

## Verification Checklist
- [ ] Added or updated automated tests
- [ ] Mutation check passed (test failed when feature logic was disabled)
- [ ] Ran `go test -race ./...` (100% pass)
- [ ] Ran `npm test` and `npm run build` in `/web` (100% pass)
- [ ] Preserved pure-Go zero CGO compatibility (`CGO_ENABLED=0`)
- [ ] Documented in `CHANGELOG.md`
