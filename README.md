# exist.io

**An open-source, self-hosted alternative to Cursor Origin—a Git forge for people and coding agents.**

exist.io is being built for teams who want their repositories, collaboration workflows, and agent integrations on infrastructure they control. The goal is a straightforward forge that works for humans first and gives coding agents a clear, permission-aware way to work with code.

> [!IMPORTANT]
> exist.io is in early development. It is not ready for production use, and there are no stable releases or supported installation instructions yet.

## What we're building

- Self-hosted Git repositories with familiar clone, fetch, and push workflows.
- Repository collaboration: pull requests, reviews, checks, and access controls.
- An API and agent-friendly integrations for coding tools and automation.
- A small, understandable deployment footprint that you can run and operate yourself.

These are project goals, not a claim that every feature is available today. Interfaces and storage are still evolving.

## Project status

The codebase is an early Go implementation. Repository and namespace metadata, grants, Git object storage, and several repository content/ref handlers are under development. The server is not yet wired as a complete forge, Git-over-HTTP transport is not available, and many collaboration and agent features remain on the roadmap.

There is no quick-start command yet: setup, first-user bootstrap, and production deployment are still being built. Check back as the project reaches its first runnable milestone.

## Roadmap

1. Wire the server, health endpoint, persistent storage, and safe first-admin setup.
2. Add authenticated Git-over-HTTP clone, fetch, and push.
3. Complete local identity, credentials, repository permissions, and core API behavior.
4. Build pull requests, reviews, checks, and repository policy workflows.
5. Add agent integrations and a compatible command-line experience.

## Technology

- Go
- SQLite for forge metadata
- `go-git` for Git object storage

The design aims to keep the core forge self-hosted and independent of Cursor services.

## Project name

**exist.io** is a home for code and collaboration that you can run on your own terms.
