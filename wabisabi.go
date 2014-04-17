package floop2

type Individual interface{}

type ScoredIndividual struct {
	Score      float64
	Individual Individual
}

type Population []ScoredIndividual

type State map[string]interface{}

type ScoringFunction func(Individual) float64

type IndividualSelection func(Population) ScoredIndividual

type Operator func([]ScoredIndividual) []ScoredIndividual
type Operators []Operator

type OperatorSelection func(Operators) Operator

type Monitor func(State, Population) (State, bool)

type Propagator func(State, Population) Population

func MakePropagator(scoring ScoringFunction, individualSelection IndividualSelection, operatorSelection OperatorSelection, operators Operators) Propagator {
	return
}

func Evolve(scoring ScoringFunction, individualSelection IndividualSelection, operatorSelection OperatorSelection, monitor Monitor, initialPopulation Population, initialState State, operators Operators) (State, Population) {
	generation := func(state State, population Population) (State, Population) {

		return
	}
	return
}
