# DCK version

This directory contains the construction-kit version of viva_tcb. The original Go sources are preserved at their original paths (revision `7b88c0d9901c58d9c610cfffe02415f6c73bebe3`), with small asset accessors so both versions use the same embedded resources.

Run the original with `go run ./cmd/vivatcb` and this version with `go run ./dck/cmd/vivatcb` from the repository root.

The choreography and assets remain in this repository. Reusable rendering and
effects come from the published `github.com/olivierh59500/democonstructionkit`
module pinned in `go.mod`. Go downloads the dependencies automatically, including
`github.com/olivierh59500/ym-player v1.0.0` for YM playback. Second Reality retains its original ST3 music synchronization.
