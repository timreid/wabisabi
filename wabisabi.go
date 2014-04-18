package wabisabi

type Individual interface{}

type ScoredIndividual struct {
	Score      float64
	Individual Individual
}

type Population []ScoredIndividual

type State map[string]interface{}

type ScoringFunction func(Individual) float64

type IndividualSelection func(Population) ScoredIndividual

type Operator func(IndividualSelection, Population, ScoringFunction) []ScoredIndividual

type Operators []Operator

type OperatorSelection func(Operators) Operator

type Monitor func(State, Population) (State, bool)

type Propagator func(State, Population) Population

func MakePropagator(scoring ScoringFunction, individualSelection IndividualSelection, operatorSelection OperatorSelection, operators Operators) Propagator {
	propagator := func(state State, population Population) Population {
		populationSize := state["populationSize"].(int)
		eliteGroupSize := state["eliteGroupSize"].(int)
		k := populationSize - eliteGroupSize
		newPopulation := make(Population, populationSize)
		//create children until the newPopulation is full
		for j := 0; j < k; {
			//select operator
			chosenOperator := operatorSelection(operators)
			//create variable number of children
			children := chosenOperator(individualSelection, population, scoring)
			//copy chlidren into pop (copy is smart about not overfilling)
			taken := copy(newPopulation, children)
			j += taken
		}
		copy(newPopulation, population[0:eliteGroupSize])
		return newPopulation
	}
	return propagator
}

func Evolve(scoring ScoringFunction, individualSelection IndividualSelection, operatorSelection OperatorSelection, monitor Monitor, initialPopulation Population, initialState State, operators Operators) (State, Population) {
	generation := func(state State, population Population) Population {

		return population
	}
	state := initialState
	population := initialPopulation
	done := false
	for !done {
		population = generation(state, population)
		state, done = monitor(state, population)
	}
	return state, population
}
