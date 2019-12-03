package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

//build workerworld correct
func buildWorkerWorld(world [][]byte, workerHeight, imageHeight, imageWidth, totalThreads, currentThreads int) [][] byte{
	workerWorld := make([][]byte, workerHeight + 2)
	for j := range workerWorld {
		workerWorld[j] = make([]byte, imageWidth)
	}

	if currentThreads == 0{
		for x := 0; x < imageWidth; x++ {
			workerWorld[0][x]=world[imageHeight - 1][x]
		}
	}else{
		for x := 0; x < imageWidth; x++ {
			workerWorld[0][x]=world[currentThreads * workerHeight - 1][x]
		}
	}

	for y := 1; y <= workerHeight; y++ {
		for x := 0; x < imageWidth; x++ {
			workerWorld[y][x]=world[currentThreads * workerHeight + y - 1][x]
		}
	}

	if currentThreads == totalThreads - 1{
		for x := 0; x < imageWidth; x++ {
			workerWorld[workerHeight+1][x]=world[0][x]
		}
	}else {
		for x := 0; x < imageWidth; x++ {
			workerWorld[workerHeight+1][x]=world[(currentThreads+1)*workerHeight][x]
		}
	}
	return workerWorld
}

// worker function
func worker(workerChan chan byte, imageHeight int, imageWidth int,out chan byte){
	// Created a new world  the channel
	world := make([][]byte, imageHeight + 2)
	for i := range world {
		world[i] = make([]byte, imageWidth)
	}
	for y := 0; y < imageHeight;y++{
		for x := 0; x < imageWidth; x++{
			world[y][x] =<- workerChan
		}
	}

	tempWorld := make([][]byte, imageHeight + 2)
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
					if world[y+i][(x+j+imageWidth)%imageWidth] != 0 {
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
	for y := 0; y < imageHeight;y++{
		for x := 0; x < imageWidth; x++{
			out <- tempWorld[y + 1][x]
		}
	}
}

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p golParams, d distributorChans, alive chan []cell, key chan rune) {

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
			val := <-d.io.inputVal
			if val != 0 {
				fmt.Println("Alive cell at", x, y)
				world[y][x] = val
			}
		}
	}

	// created a 2s ticker.
	ticker := time.NewTicker(2000 * time.Millisecond)

	// Calculate the new state of Game of Life after the given number of turns.
	for turns := 0; turns < p.turns; turns++ {
		select {
		// case <-ticker.C:
		case c := <- key:
			if c == 's' {
				printBoard(d, p, world,turns)
			} else if c == 'q' {
				printBoard(d, p, world, turns)
				fmt.Println("Terminated.")
				return
			} else if c == 'p' {
				fmt.Println(turns)
				fmt.Println("pausing")
				for {
					tempKey := <-key
					if tempKey == 'p' {
						fmt.Println("continuing.")
						break
					}
				}
			}
		case <-ticker.C:
			var finalAlive []cell
			// Go through the world and append the cells that are still alive.
			for y := 0; y < p.imageHeight; y++ {
				for x := 0; x < p.imageWidth; x++ {
					if world[y][x] != 0 {
						finalAlive = append(finalAlive, cell{x: x, y: y})
					}
				}
			}
			fmt.Println("number of alive cells is:", len(finalAlive))

		default:
		}

		//Put logic outside of select such that when other cases run, the logic doesn't skip a turn.
		workerHeight := float32(p.imageHeight) / float32(p.threads)
		out := make([] chan byte, p.threads)

		for i := 0; i < p.threads; i++ {
			out[i] = make(chan byte)
			workerChan := make(chan byte)
			endY := int(float32(i+1)*workerHeight)
			startY := int(float32(i)*workerHeight)
			intWorkerHeight := endY - startY
			fmt.Println("started i = ", i)
			//build slices the workers need to work on
			workerWorld := buildWorkerWorld(world, intWorkerHeight, p.imageHeight, p.imageWidth, p.threads, i)
			go worker(workerChan, intWorkerHeight+2, p.imageWidth, out[i])
			//Send world cells to workers
			for y := 0; y < intWorkerHeight+2; y++ {
				for x := 0; x < p.imageWidth; x++ {
					workerChan <- workerWorld[y][x]
				}
			}
		}
		for i := 0; i < p.threads; i++ {
			endY := int(float32(i+1) * workerHeight)
			startY := int(float32(i) * workerHeight)
			intWorkerHeight := endY - startY
			//slices from workers
			tempOut := make([][]byte, intWorkerHeight)
				for i := range tempOut {
					tempOut[i] = make([]byte, p.imageWidth)
				}
				for y := 0; y < intWorkerHeight; y++ {
					for x := 0; x < p.imageWidth; x++ {
						tempOut[y][x] = <-out[i]
					}
				}
				//println("tempOut  i=",i)
				for y := 0; y < intWorkerHeight; y++ {
					for x := 0; x < p.imageWidth; x++ {
						//print(tempOut[y+1][x])
						world[i*intWorkerHeight+y][x] = tempOut[y][x]
					}
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

		d.io.command <- ioCheckIdle
		<-d.io.idle
		// Return the coordinates of cells that are still alive.
		alive <- finalAlive
	}

func printBoard(d distributorChans, p golParams, world[][]byte, turn int){
	d.io.command <- ioOutput
	d.io.filename <- strings.Join([]string{strconv.Itoa(p.imageWidth), strconv.Itoa(p.imageHeight), strconv.Itoa(turn)}, "x")
	d.io.outputWorld <- world
}