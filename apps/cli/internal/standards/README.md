# Pinned v1.0 sources

The two files in `assets/` are byte-for-byte copies of the user-supplied `AI-Multi-Platform-Engineering-Standard-v1.0.zip` and `AI-Product-Visual-Design-Standard-v1.0.zip`. They are embedded into the CLI and unpacked under each new project's `.aistd/standards/` directory. The engineering archive includes a reference app; the visual archive includes a concept image. Both remain reference material in that managed subtree.

| Archive | SHA256 |
| --- | --- |
| `engineering-v1.0.zip` | `6cad0f6013e8ce7c925041b9d7ad8913d404e184e2828d711bf03edbaadf6cc6` |
| `visual-v1.0.zip` | `f18ea62a1cb552a43b31e560d03519757523dab819784748eb597bb1b8c757e5` |

`standards.Load` checks these pinned archive hashes. Generated projects record each unpacked file's SHA256 in `.aistd/standards.lock.json`; `doctor` compares the lock and all managed files against the CLI's embedded sources. This checks snapshot integrity, not publisher identity or product conformance. A future standard revision needs new source archives, pinned hashes, and explicit migration behavior; do not silently replace v1.0 bytes.
