package wabisabi

import "math"
import "math/rand"
import "fmt"
import "sort"

type State map[string]interface{}

type Genotype interface{}

type Individual struct {
	Score  float64
	State  State
	Genome Genotype
}

type Population []Individual

func (p Population) Len() int {
	return len(p)
}

func (p Population) Less(i, j int) (x bool) {
	tempIndI := p[i]
	tempIndJ := p[j]
	if tempIndI.Score < tempIndJ.Score || math.IsNaN(tempIndI.Score) {
		x = false
	} else {
		x = true
	}
	return x
}

func (p Population) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
	return
}

type Fitness func(Genotype) (float64, State)

type Selection func(Population) Individual

type Operator struct {
	P  func(State) float64
	Op func(Fitness, Selection, Population) Population
}

type Monitor func(State, Population) (State, bool)

func SimpleMonitor(state State, population Population) (newState State, done bool) {
	generation := state["generation"].(int)
	max := state["maxFitness"].(float64)
	goal := state["goalFitness"].(float64)
	sum := 0.0
	for _, c := range population {
		//min, max, mean, std dev
		cs := c.Score

		if cs > max {
			max = cs
		}
		if !math.IsNaN(cs) {
			sum = sum + cs
		}
	}
	newState = state
	mean := sum / float64(len(population))
	newState["generation"] = generation + 1
	newState["maxFitness"] = max
	newState["meanFitness"] = mean
	fmt.Printf("%4d\t%4.3f\t%4.3f", generation, max, mean)
	fmt.Println()
	if max >= goal {
		done = true
	}
	return
}

func Evolve(fitness Fitness, selection Selection, monitor Monitor, operators []Operator, initialPopulation Population, initialState State) (State, Population) {

	state := initialState
	state["generation"] = 0
	population := initialPopulation
	done := false

	operatorSelection := func(operators []Operator) Operator {
		i := 0
		x := rand.Float64()
		for P := operators[i].P(state); x > P; P = operators[i].P(state) {
			x = x - P
			i++
		}
		return operators[i]

	}

	propagate := func(state State, population Population) Population {
		populationSize := state["populationSize"].(int)
		eliteGroupSize := state["eliteGroupSize"].(int)
		k := populationSize - eliteGroupSize
		newPopulation := *new(Population)
		//create children until the newPopulation is full
		for j := 0; j < k; {
			//select operator
			chosenOperator := operatorSelection(operators)
			//create some variable number of children
			children := chosenOperator.Op(fitness, selection, population)
			nChildren := len(children)
			//fmt.Println("children: ", children)
			newPopulation = append(newPopulation, children...)
			j += nChildren
		}
		//copy elite into pop
		newPopulation = append(newPopulation[0:k], population[0:eliteGroupSize]...)
		sort.Sort(newPopulation)
		return newPopulation
	}

	for !done {
		population = propagate(state, population)
		state, done = monitor(state, population)
	}

	return state, population
}
