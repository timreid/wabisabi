package wabisabi

import "math"
import "math/rand"
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

type Operators []Operator

type Monitor func(State, Population) (State, bool)

//simple exponential selection with las vegas resampling
func MakeSelection(scale float64) Selection {
	selection := func(population Population) (selected Individual) {
		k := len(population)
		sigma := scale * float64(k)
		f := func() (i int) { i = int(math.Floor(math.Abs(rand.NormFloat64() * sigma))); return }
		sample := 0
		//las vegas resampling
		for sample = f(); sample >= k; sample = f() {
		}
		selected = population[sample]
		return
	}
	return selection
}

func Evolve(fitness Fitness, selection Selection, monitor Monitor, operators Operators, initialPopulation Population, initialState State) (State, Population) {

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
		//ugly corner case: elite group can be bigger than the population
		eliteGroupSize := int(math.Min(float64(state["eliteGroupSize"].(int)), float64(len(population))))

		//how many children do we have to create
		k := populationSize - eliteGroupSize

		//create children until the newPopulation is full
		newPopulation := *new(Population)
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
		newPopulation = append(newPopulation, population[0:eliteGroupSize]...)
		sort.Sort(newPopulation)
		return newPopulation
	}

	//main loop repeats until monitor tells us to stop
	for !done {
		population = propagate(state, population)
		state, done = monitor(state, population)
	}

	return state, population
}
