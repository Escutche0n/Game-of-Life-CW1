package main

import (
	"fmt"
	"strconv"
	"strings"
)

func worker(height int, width int, world chan byte, out chan<- [][]byte) {
	//Create the 2D slice to store tempWorld
	tempWorld := make([][]byte, height)
	for i := range tempWorld {
		tempWorld[i] = make([]byte, width)
	}

	for x := 0; x < width; x++ {
		tempWorld[0][x] = <-world
		tempWorld[height][x] = <-world
	}
	for y := 1; y < height-1; y++ {
		for x := 0; x < width; x++ {
			tempWorld[y][x] = <-world
		}
	}

	for y := 1; y < height-1; y++ {
		for x := 0; x < width; x++ {
			neighboursAlive := 0
			for i := -1; i < 2; i++ {
				for j := -1; j < 2; j++ {
					if i == 0 && j == 0 {
						continue
					}
					if tempWorld[(y+i+height)%height][(x+j+width)%width] != 0 {
						neighboursAlive++
					}
				}
			}
			if tempWorld[y][x] == 255 {
				if (neighboursAlive < 2) || (neighboursAlive > 3) {
					tempWorld[y][x] = 0
				} else {
					tempWorld[y][x] = 255
				}
			}
			if tempWorld[y][x] == 0 {
				if neighboursAlive == 3 {
					tempWorld[y][x] = 255
				} else {
					tempWorld[y][x] = 0
				}
			}
		}
	}
	for y := 1; y < height-1; y++ {
		for x := 0; x < width; x++ {
			out <- tempWorld
		}
	}
}

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p golParams, d distributorChans, alive chan []cell) {

	// Create the 2D slice to store the world.
	world := make([][]byte, p.imageHeight)
	for i := range world {
		world[i] = make([]byte, p.imageWidth)
	}

	// 新建一个 workHeight 存储每个 worker 的高度
	workerHeight := p.imageHeight / p.threads

	// 新建了一个类型为 [][] byte 的 output channel
	output := make(chan<- [][]byte, p.imageHeight)

	// 新建了一个类型为 [] byte 的 workers channel
	workers := make([]chan byte, p.threads)
	for i := range workers {
		workers[i] = make(chan byte, workerHeight+2)
		go worker((workerHeight+2+p.imageHeight)%p.imageHeight, p.imageWidth, workers[i], output)
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
				workers[p.threads] <- world[y][x]
			}
		}
		for i := 0; i < p.threads; i++ {
			for x := 0; x < p.imageWidth; x++ {
				workers[i] <- world[(i*workerHeight-1+p.imageHeight)%p.imageHeight][x]
				workers[i] <- world[((i+1)*workerHeight+p.imageHeight)%p.imageHeight][x]
			}
		}
		for y := 0; y < p.imageHeight; y++ {
			for x := 0; x < p.imageWidth; x++ {
				world[y][x] = <-workers[y/workerHeight]
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

	// Make sure that the Io has finished any output before exiting.
	d.io.command <- ioCheckIdle
	<-d.io.idle

	// Return the coordinates of cells that are still alive.
	alive <- finalAlive
}
