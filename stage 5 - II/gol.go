package main

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
)

func buildWorkerWorld(world [][]byte, workerHeight, imageHeight, imageWidth, currentThreads int) [][] byte{
	workerWorld := make([][]byte, workerHeight + 2)
	for j := range workerWorld {
		workerWorld[j] = make([]byte, imageWidth)
	}

	for x := 0; x < imageWidth; x++ {
		workerWorld[0][x]=world[(currentThreads * workerHeight + imageHeight - 1) % imageHeight][x]
	}

	for y := 1; y <= workerHeight; y++ {
		for x := 0; x < imageWidth; x++ {
			workerWorld[y][x]=world[currentThreads * workerHeight + y - 1][x]
		}
	}

	for x := 0; x < imageWidth; x++ {
		workerWorld[workerHeight+1][x]=world[((currentThreads+1)*workerHeight + imageHeight)%imageHeight][x]
	}
	return workerWorld
}

// define a mutex.
var mutex = &sync.Mutex{}

// worker function
func worker(world [][]byte, tempWorld [][]byte, imageHeight int, imageWidth int, numOfWorker int, done func()){
	defer done()
	tempWorld = make([][]byte, imageHeight + 2)
	for i := range world {
		tempWorld[i] = make([]byte, imageWidth)
	}

	for y := 1; y <= imageHeight; y++ {
		for x := 0; x < imageWidth; x++ {
			var neighboursAlive = 0

			for i := -1; i < 2; i++ {
				for j := -1; j < 2; j++ {
					// Mark all of the neighbours excluding the cell itself.
					if i == 0 && j == 0 {
						continue
					}
					// If the cell is on the edge of the diagram, mod it to fix the rule of the game.
					if world[y + i][(x + j + imageWidth) % imageWidth] != 0 {
						neighboursAlive += 1
					}
				}
			}
			if world[y][x] == 255 {
				// If less than 2 or more than 3 neighbours, live cells dead.
				if (neighboursAlive < 2) || (neighboursAlive > 3) {
					tempWorld[y][x] = 0
				} else {
					tempWorld[y][x] = 255
				}
			}

			// When the colour is black, the cell status is dead, parameter is 0.
			if world[y][x] == 0 {
				// If 3 neighbours alive, dead cells alive.
				if neighboursAlive == 3 {
					tempWorld[y][x] = 255
				} else {
					tempWorld[y][x] = 0
				}
			}
		}
	}

	mutex.Lock()
	for y := 1; y < imageHeight+1; y++{
		world[imageHeight*numOfWorker+y-1] = tempWorld[y] //access to memory
	}
	mutex.Unlock()

}

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p golParams, d distributorChans, alive chan []cell) {

	// Create the 2D slice to store the world.
	world := make([][]byte, p.imageHeight)
	for i := range world {
		world[i] = make([]byte, p.imageWidth)
	}

	// Request the io goroutine to read in the image with the given filename.
	d.io.command <- ioInput
	d.io.filename <- strings.Join([]string{strconv.Itoa(p.imageWidth), strconv.Itoa(p.imageHeight)}, "x")

	// The io goroutine sends the requested image byte by byte, in rows.
	for y := 0; y < p.imageHeight; y++ {
		for x := 0; x < p.imageWidth; x++ {
			val := <- d.io.inputVal
			if val != 0 {
				fmt.Println("Alive cell at", x, y)
				world[y][x] = val
			}
		}
	}
	workerHeight := p.imageHeight / p.threads
	// Calculate the new state of Game of Life after the given number of turns.
	for turns := 0; turns < p.turns; turns++ {

		workerWorld := make([][]byte,p.imageHeight)
		for i := range workerWorld{
			workerWorld[i] = make([]byte, p.imageWidth)
		}

		copy(workerWorld,world)

		var wg sync.WaitGroup
		wg.Add(p.threads)                                                    //using counter count the num of worker
		for i := 0; i < p.threads; i++{

			go   worker(world, buildWorkerWorld(workerWorld, workerHeight, p.imageHeight, p.imageWidth, p.threads),p.imageHeight, p.imageWidth,i,wg.Done)   //in workers they give the new world

		}
		wg.Wait()                                                       //wait for workers
		//receive world parts and re-gather them

		world = workerWorld
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