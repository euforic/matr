# Security

This repository is a local developer tool, but it still executes shell commands and compiles generated code. That means changes should be reviewed with these risks in mind:

- command execution paths should stay explicit
- filesystem writes should stay scoped to the working tree and cache directory
- generated code paths should remain predictable and documented
- untrusted input should not be executed without clear intent

When changing shell execution, file creation, or generated wrapper behavior, document the risk in the same change.
