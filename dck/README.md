# DCK version

This directory contains the construction-kit version of viva_tcb. The original Go sources are preserved at their original paths (revision `7b88c0d9901c58d9c610cfffe02415f6c73bebe3`), with small asset accessors so both versions use the same embedded resources.

Run the original with `go run ./cmd/vivatcb` and this version with `go run ./dck/cmd/vivatcb` from the repository root.

The choreography and assets stay local; reusable rendering and effects live in `../../lib/democonstructionkit`. Second Reality retains its original ST3 music synchronization.
