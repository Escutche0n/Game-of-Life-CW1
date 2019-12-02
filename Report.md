# Report (VIVA FRIDAY 14:30 )

##### Xin Ye Elvis Chen



### Functionality and Design

The overall task is to design and implement a concurrent, multi-threaded GO program which simulates the Game of Life on an image matrix. The game matrix should be initialised from a PGM (Portable Gray Map) image and the user should be able to export the game matrix as PGM files.

The Game of Life logic is successfully implemented as it was described in the task introductionin stage 1a working as a single-threaded program. Based on the skeleton provided, most of the code were modified in the `distributor` function in  `gol.go`. For each turn, the number of neighbours was initialised to 0. Meanwhile, all the neighbours excepting the centre-cell were marked. To comply the mission, the coordinates were looped from `-1` to `1` that neighbours of non-edging cells can be all located. Whereas the edging cells were using **modulo** operation, adding the whole column or row to avoid out-of-range issue. The centre-cell were skipped with `continue` when both `i`and `j` were `0`. After the number of neighbours were counted, if-statements were executed to enact the whole logic. All of the generated cells were stored in a placeholder `tempWorld`, which will be replaced by world at the end of it.

stage 1b use multi-threads. we implement a buildWorkerWorld to 











### Tests, Experiments and Critical Analysis (3 pages max)

