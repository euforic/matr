# Product Sense

`matr` gives Go projects a small task runner that uses exported Go functions as commands. It is for developers who want task definitions to stay in Go instead of moving into shell scripts or separate build DSLs.

The core user promise is straightforward:

- point `matr` at a `Matrfile`
- discover exported task functions
- expose those tasks as CLI commands
- execute them with predictable output and failure behavior

Changes should preserve that simplicity. New harness or repo-process work should stay lightweight and should not add product complexity to the runtime unless the runtime itself benefits.
