package main

import (
	"fmt"
	"strconv"
	"strings"
)

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p golParams, d distributorChans, alive chan []cell) {

	// Create the 2D slice to store the world.
	world := make([][]byte, p.imageHeight)
	for i := range world {
		world[i] = make([]byte, p.imageWidth)
	}

	// Created another 2D slice to store the world that has cache.
	nextWorld := make([][]byte, p.imageHeight)
	for i := range nextWorld {
		nextWorld[i] = make([]byte, p.imageWidth)
	}

	// Request the io goroutine to read in the image with the given filename.
	d.io.command <- ioInput
	d.io.filename <- strings.Join([]string{strconv.Itoa(p.imageWidth), strconv.Itoa(p.imageHeight)}, "x")

	// The io goroutine sends the requested image byte by byte, in rows.
	for y := 0; y < p.imageHeight; y++ {
		for x := 0; x < p.imageWidth; x++ {
			val := <-d.io.inputVal
			if val != 0 {
				fmt.Println("Alive cell at", x, y)
				world[y][x] = val
			}
		}
	}

	// Calculate the new state of Game of Life after the given number of turns.
	for turns := 0; turns < p.turns; turns++ {
		for y := 0; y < p.imageHeight; y++ {
			for x := 0; x < p.imageWidth; x++ {
				// Placeholder for the actual Game of Life logic: flips alive cells to dead and dead cells to alive.
				// Initialise the neighboursAlive to 0.
				neighboursAlive := 0
				for i := -1; i < 2; i++ {
					for j := -1; j < 2; j++ {
						// Mark all of the neighbours excluding the cell itself.
						if i == 0 && j == 0 {
							continue
						}
						// If the cell is on the edge of the diagram, mod it to fix the rule of the game.
						if world[(y+i+p.imageHeight)%p.imageHeight][(x+j+p.imageWidth)%p.imageWidth] != 0 {
							neighboursAlive += 1
						}
					}
				}

				// When the colour is white, the cell status is alive, parameter is 255.
				if world[y][x] == 255 {
					// If less than 2 or more than 3 neighbours, live cells dead.
					if (neighboursAlive < 2) || (neighboursAlive > 3) {
						nextWorld[y][x] = 0
					} else {
						nextWorld[y][x] = 255
					}
				}

				// When the colour is black, the cell status is dead, parameter is 0.
				if world[y][x] == 0 {
					// If 3 neighbours alive, dead cells alive.
					if neighboursAlive == 3 {
						nextWorld[y][x] = 255
					} else {
						nextWorld[y][x] = 0
					}
				}
			}
		}
		for y := 0; y < p.imageHeight; y++ {
			for x := 0; x < p.imageWidth; x++ {
				// Replace placeholder nextWorld[y][x] with the real world[y][x]
				world[y][x] = nextWorld[y][x]
			}
		}
	}

	// Create an empty slice to store coordinates of cells that are still alive after p.turns are done.
	var finalAlive []cell
	// Go through the world and append the cells that are still alive.
	for y := 0; y < p.imageHeight; y++ {
		for x := 0; x < p.imageWidth; x++ {
			if world[y][x] != 0 {
				finalAlive = append(finalAlive, cell{x: x, y: y})
			}
		}
	}

	// Request the io goroutine to write in the image with the given filename.
	d.io.command <- ioOutput
	d.io.filename <- strings.Join([]string{strconv.Itoa(p.imageWidth), strconv.Itoa(p.imageHeight), strconv.Itoa(p.turns)}, "x")

	// Send the world to finalBoard
	d.io.finalBoard <- world

	// Make sure that the Io has finished any output before exiting.
	d.io.command <- ioCheckIdle
	<-d.io.idle

	// Return the coordinates of cells that are still alive.
	alive <- finalAlive
}
