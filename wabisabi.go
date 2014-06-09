package wabisabi

import "math"
import "math/rand"
import "sort"
import "fmt"

//used to hold metadeta for individuals and the state of the simulation
type Meta map[string]interface{}

//an abstract representation of the genome of an individual
type Genotype interface{}

//an individual
type Individual struct {
	Genome Genotype
	Score  float64
	Meta   Meta
}

//populations can be sorted by score
type Population []Individual

func (p Population) Len() int {
	return len(p)
}

func (p Population) Less(i, j int) (x bool) {
	A := p[i]
	B := p[j]
	return A.Score > B.Score
}

func (p Population) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
	return
}

//assigns a score and metadeta to a genotype
type Fitness func(Genotype) (float64, Meta)

//selects an individual from the population for reproduction
type Selection func(Population) Individual

//an evolutionary operator with associated propability of selection for use in reproduction
type Operator struct {
	//P  func(Meta) float64
	P  float64
	Op func(Fitness, Selection, Population) Population
}

type Operators []Operator

//responsible for logging and ending the experiment
type Monitor func(Meta, Population) (Meta, bool)

func Evolve(fitness Fitness, selection Selection, monitor Monitor, operators Operators, initialPopulation Population, initialState Meta) (Meta, Population) {

	state := initialState

	operatorSelection := func(operators []Operator) Operator {
		i := 0
		x := rand.Float64()
		for P := operators[i].P; x > P; P = operators[i].P {
			x = x - P
			i++
		}
		return operators[i]

	}

	propagate := func(state Meta, population Population) Population {
		populationSize := state["populationSize"].(int)
		//ugly corner case: elite group might be bigger than the population
		eliteGroupSize := int(math.Min(float64(state["eliteGroupSize"].(int)), float64(len(population))))

		//how many children do we have to create
		k := populationSize - eliteGroupSize

		//create children until the newPopulation is full
		newPopulation := Population{}
		for j := 0; j < k; {
			//select operator
			chosenOperator := operatorSelection(operators)
			//create some variable number of children
			children := chosenOperator.Op(fitness, selection, population)
			nChildren := len(children)
			newPopulation = append(newPopulation, children...)
			j = j + nChildren
		}
		//copy elite into pop
		elites := population[0:eliteGroupSize]
		newPopulation = append(newPopulation, elites...)
		sort.Sort(newPopulation)
		return newPopulation
	}

	state["generation"] = 0
	population := initialPopulation
	done := false
	fmt.Println("simulation starting")

	//main loop repeats until monitor tells us to stop
	monitor(state, population)
	for !done {
		population = propagate(state, population)
		state, done = monitor(state, population)
	}

	return state, population
}
