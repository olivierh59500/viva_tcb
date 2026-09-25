# DCK version

This directory contains the construction-kit version of viva_tcb. The original Go sources are preserved at their original paths (revision `7b88c0d9901c58d9c610cfffe02415f6c73bebe3`), with small asset accessors so both versions use the same embedded resources.

Run the original with `go run ./cmd/vivatcb` and this version with `go run ./dck/cmd/vivatcb` from the repository root.

The choreography and assets remain in this repository. Reusable rendering and
effects come from the published `github.com/olivierh59500/democonstructionkit`
module pinned in `go.mod`. Music is opened with `sound.Open`; DCK selects the decoder from the asset and
provides the configured stereo PCM format. The demo keeps its playback level and loop settings.

The moving tile is now a `composite.RotozoomBackground` driven by the editable
`presets.VivaRotozoom` program. Entrance motion, orbit, zoom, rotation and
texture phase advance independently. The GPU renderer repeats the source tile
with one quad and no giant pre-tiled image. A 20-second comparison matched all
1,200 decoded frames through the entrance and rotozoom stages.

For silent, deterministic native frames of the DCK version:

```sh
go run ./dck/cmd/capture -frames 0,1,60,240,600,1200 -out captures/viva
```

The four pseudo-3D scrolltexts now use one `scrolling.Config.Pseudo3D` effect.
Their text, font metrics, depth harmonics, spacing, mirrored scale and clipping
remain editable parameters. Eight captures through frame 4,800 match the
preceding DCK renderer in every color and alpha channel.
The ten moving DMA logos now use `sprites.RecurrentFormation`. Four harmonic
seeds produce all poses with editable count, phase spacing and vertical
amplitude. Eight captures through frame 4,800 remain identical in every channel.
The moving title uses `composite.RasterTitle` in its small-canvas mode. Its
two raster clocks and third repeated strip remain editable; fourteen captures
at wrap and title-cue boundaries remain identical in every channel.
