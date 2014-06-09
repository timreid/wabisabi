package utilities

import "github.com/timreid/wabisabi"
import "fmt"
import "os"
import "math/rand"
import "math"
import "os/signal"
import "syscall"
import "runtime"
import "io"

//simple exponential selection with las vegas resampling
func MakeExponentialSelection(scale float64) (selection wabisabi.Selection) {
	selection = func(population wabisabi.Population) (selected wabisabi.Individual) {
		k := len(population)
		sigma := scale * float64(k)
		//single tailed gaussian with std dev of sigma
		sample := func() (i int) { i = int(math.Floor(math.Abs(rand.NormFloat64() * sigma))); return }
		x := 0
		//las vegas resampling
		for x = sample(); x >= k; x = sample() {
		}
		selected = population[x]
		return
	}
	return
}

func MakeSimpleMonitor(writer io.Writer) wabisabi.Monitor {
	//setup interrupt handler
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)

	monitor := func(state wabisabi.Meta, population wabisabi.Population) (newMeta wabisabi.Meta, done bool) {
		generation := state["generation"].(int)
		goal := state["goalFitness"].(float64)
		maxGenerations := state["maxGenerations"].(int)

		newMeta = state
		newMeta["generation"] = generation + 1

		elite := population[0]

		itemCount := elite.Meta["itemCount"].(int)
		weight := elite.Meta["weight"].(float64)

		fmt.Fprintf(writer, "%d\t%f\t%d\t%f\n", generation, elite.Score, itemCount, weight)

		if elite.Score >= goal || generation > maxGenerations {
			done = true
		}

		//also done if we have an interrupt
		runtime.Gosched() //give the interrupt handler a chance to run
		select {
		case _ = <-c:
			done = true
		default:
			//carry on my wayward son
		}
		return
	}
	return monitor
}
