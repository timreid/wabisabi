package main

import "fmt"
import "github.com/timreid/wabisabi"
import "github.com/timreid/wabisabi/utilities"
import "os"
import "math/rand"
import "sort"
import "math"

//an item which can be stolen
type Goody struct {
	Value  float64
	Weight float64
	VPW    float64
}

type Goodies []Goody

//the solution is a choice of goodies
type Genotype []bool

func (genotype Genotype) String() (output string) {
	k := len(genotype)
	x := ""
	for i := 0; i < k; i++ {
		if genotype[i] {
			x = "1"
		} else {
			x = "0"
		}
		output = output + x
	}
	return output
}

func MakeRandomGoodies(n int) (goodies Goodies) {
	for i := 0; i < n; i++ {
		value := 100.0 * math.Abs(rand.NormFloat64())
		weight := math.Abs(rand.NormFloat64())
		vpw := value / weight
		newGoodie := Goody{value, weight, vpw}
		goodies = append(goodies, newGoodie)
	}
	return
}

//split each parent at random point, take one part from each parent
func SinglePointCrossover(fitness wabisabi.Fitness, selection wabisabi.Selection, population wabisabi.Population) (children wabisabi.Population) {
	//select 2 parents
	mom := selection(population)
	momGenome := mom.Genome.(Genotype)
	dad := selection(population)
	dadGenome := dad.Genome.(Genotype)

	//assumption: mom and dad are the same length
	k := len(momGenome)

	childGenome := make(Genotype, k)

	//choose crossover location
	cut := rand.Intn(k)

	//for each point, take gene from either mom or dad
	for i := range momGenome {
		if i < cut {
			childGenome[i] = momGenome[i]
		} else {
			childGenome[i] = dadGenome[i]
		}
	}

	childScore, childMeta := fitness(childGenome)
	child := wabisabi.Individual{Score: childScore, Meta: childMeta, Genome: childGenome}
	children = wabisabi.Population{child}
	return
}

// //an interval is chosen from one parent, rest is taken from the other
// func TwoPointCrossover(fitness wabisabi.Fitness, selection wabisabi.Selection, population wabisabi.Population) (children wabisabi.Population) {
// 	//select 2 parents
// 	mom := selection(population)
// 	momGenome := mom.Genome.(Genotype)
// 	dad := selection(population)
// 	dadGenome := dad.Genome.(Genotype)

// 	//assumption: mom and dad are the same length
// 	k := len(momGenome)

// 	//allocate
// 	//childGenome := make(Genotype, k)

// 	//choose crossover location
// 	rightCut := 1 + rand.Intn(k-1)
// 	leftCut := rand.Intn(rightCut)

// 	childGenome := make(Genotype, k)
// 	for i := 0; i < k; i++ {
// 		if i < leftCut || i >= rightCut {
// 			childGenome[i] = momGenome[i]
// 		} else {
// 			childGenome[i] = dadGenome[i]
// 		}
// 	}
// 	//fmt.Printf("\n%v\n%v\n%v\n\n", momGenome, childGenome, dadGenome)
// 	childScore, childMeta := fitness(childGenome)
// 	child := wabisabi.Individual{Score: childScore, Genome: childGenome, Meta: childMeta}
// 	children = wabisabi.Population{child}
// 	return
// }

func mod(x, k int) (z int) {
	z = int(math.Mod(float64(x), float64(k)))
	return
}

//remove an item at random and try to replace it with 2
func makeSplitItemMutation(goodies Goodies, maxWeight float64) (splitItemMutation func(fitness wabisabi.Fitness, selection wabisabi.Selection, population wabisabi.Population) (children wabisabi.Population)) {

	splitItemMutation = func(fitness wabisabi.Fitness, selection wabisabi.Selection, population wabisabi.Population) (children wabisabi.Population) {
		victim := selection(population)
		victimGenome := victim.Genome.(Genotype)
		k := len(victimGenome)

		newGuyGenome := make(Genotype, k)
		copy(newGuyGenome, victimGenome)
		//fmt.Println("split")
		//fmt.Println(newGuyGenome)
		//remove an item
		startingPoint := rand.Intn(k)
		removedItem := startingPoint
		for i := 0; i < k; i++ {
			x := mod(startingPoint+i, k)
			if newGuyGenome[x] {
				removedItem = x
				newGuyGenome[x] = false
				break
			}
		}
		remainingWeight := maxWeight - victim.Meta["weight"].(float64) + goodies[removedItem].Weight

		//now try to find two empty spots which we can fit
		howMany := 0
	findEm:
		for howMany < 2 {
			//find something to stuff in the bag
			startingPoint := rand.Intn(k)
			for i := 0; i < k; i++ {
				x := mod(startingPoint+i, k)
				if x != removedItem && !newGuyGenome[x] && goodies[x].Weight <= remainingWeight {
					newGuyGenome[x] = true
					howMany = howMany + 1
					continue findEm
				}
			}
			break //bummer, we could not find an item
		}

		newGuy := wabisabi.Individual{}
		newGuyScore, newGuyMeta := fitness(newGuyGenome)
		newGuy.Genome = newGuyGenome
		newGuy.Score = newGuyScore
		newGuy.Meta = newGuyMeta
		children = wabisabi.Population{newGuy}
		//fmt.Println(newGuyGenome)
		return
	}
	return
}

//idea:greedyizer
//find the item with the highest VPW which we do not currently have
//remove items of lower VPW until we have space for it
func makeGreedyAddMutation(goodies Goodies, maxWeight float64) (removeItemMutation func(fitness wabisabi.Fitness, selection wabisabi.Selection, population wabisabi.Population) (children wabisabi.Population)) {

	removeItemMutation = func(fitness wabisabi.Fitness, selection wabisabi.Selection, population wabisabi.Population) (children wabisabi.Population) {
		victim := selection(population)
		victimGenome := victim.Genome.(Genotype)
		k := len(victimGenome)

		newGuyGenome := make(Genotype, k)
		copy(newGuyGenome, victimGenome)

		//find item with highest VPW which we do not already have
		best := 0
		bestVPW := 0.0
		for key, value := range goodies {
			//if this item has a better VPW, and we don't have it, and it can fit alone in the bag (nasty case)
			if value.VPW > bestVPW && !newGuyGenome[key] && goodies[key].Weight <= maxWeight {
				best = key
				bestVPW = value.VPW
			}
		}
		bestWeight := goodies[best].Weight
		capacity := maxWeight - victim.Meta["weight"].(float64)
		//fmt.Println("best", best, "bestWeight", bestWeight, "bestVPW", bestVPW, "capacity", capacity)

		//now remove items until we have enough space
		//modular scan over knapsack
		//if we have the item, and its VPW < bestVPW
		// for key, value := range victimGenome {
		// 	if value && goodies[key].VPW < bestVPW {
		// 		victimGenome[key] = false
		// 		capacity = capacity + goodies[key].Weight
		// 		if capacity >= bestWeight {
		// 			break
		// 		}
		// 	}
		// }
		//ASSUMPTION! no goody is larger than the bag!
		//also, this is a las vegas algo
		for capacity < bestWeight {
			//fmt.Println("making space:", capacity)
			x := rand.Intn(k)
			potential := newGuyGenome[x]
			//fmt.Println("bestVPW", bestVPW, "goody VPW", goodies[x].VPW)

			if potential {
				//fmt.Println("removing")
				newGuyGenome[x] = false //drop this item
				capacity = capacity + goodies[x].Weight
				//fmt.Println("capacity", capacity)
			}
		}

		//now we have sufficient space, so take it
		//fmt.Println("item taken")
		newGuyGenome[best] = true

		//fmt.Println(newGuyGenome)

		newGuy := wabisabi.Individual{}
		newGuyScore, newGuyMeta := fitness(newGuyGenome)
		newGuy.Genome = newGuyGenome
		newGuy.Score = newGuyScore
		newGuy.Meta = newGuyMeta
		children = wabisabi.Population{newGuy}

		return
	}
	return
}

//remove an item at random
func makeRemoveItemMutation(goodies Goodies, maxWeight float64) (makeRemoveItemMutation func(fitness wabisabi.Fitness, selection wabisabi.Selection, population wabisabi.Population) (children wabisabi.Population)) {

	makeRemoveItemMutation = func(fitness wabisabi.Fitness, selection wabisabi.Selection, population wabisabi.Population) (children wabisabi.Population) {
		victim := selection(population)
		victimGenome := victim.Genome.(Genotype)
		k := len(victimGenome)

		newGuyGenome := make(Genotype, k)
		copy(newGuyGenome, victimGenome)

		//randomly choose where to start looking for an item to drop
		startingPoint := rand.Intn(k)

		//modular scan over knapsack to find item to drop
		done := false
		for i := 0; i < k && !done; i++ {
			x := mod(startingPoint+i, k)
			if newGuyGenome[x] {
				newGuyGenome[x] = false
				done = true
			}
		}
		newGuy := wabisabi.Individual{}
		newGuyScore, newGuyMeta := fitness(newGuyGenome)
		newGuy.Genome = newGuyGenome
		newGuy.Score = newGuyScore
		newGuy.Meta = newGuyMeta
		children = wabisabi.Population{newGuy}
		return
	}
	return
}

//add a random item
func makeAddItemMutation(goodies Goodies, maxWeight float64) (addItemMutation func(fitness wabisabi.Fitness, selection wabisabi.Selection, population wabisabi.Population) (children wabisabi.Population)) {

	addItemMutation = func(fitness wabisabi.Fitness, selection wabisabi.Selection, population wabisabi.Population) (children wabisabi.Population) {
		victim := selection(population)
		victimGenome := victim.Genome.(Genotype)
		k := len(victimGenome)

		newGuyGenome := make(Genotype, k)
		copy(newGuyGenome, victimGenome)

		//fmt.Println("addItemMutation")
		//fmt.Println(victimGenome)

		//randomly choose where to start looking for an item to add
		startingPoint := rand.Intn(k)
		remainingWeight := maxWeight - victim.Meta["weight"].(float64)

		//modular scan over knapsack to find item to add
		for i := 0; i < k; i++ {
			x := mod(startingPoint+i, k)
			if !newGuyGenome[x] && goodies[x].Weight <= remainingWeight {
				newGuyGenome[x] = true
				break
			}
		}
		newGuy := wabisabi.Individual{}
		newGuyScore, newGuyMeta := fitness(newGuyGenome)
		newGuy.Genome = newGuyGenome
		newGuy.Score = newGuyScore
		newGuy.Meta = newGuyMeta
		children = wabisabi.Population{newGuy}
		return
	}
	return
}

//remove an item, and replace it with an item which improves Value per Weight
func makeImproveVPWMutation(goodies Goodies, maxWeight float64) (improveVPWMutation func(fitness wabisabi.Fitness, selection wabisabi.Selection, population wabisabi.Population) (children wabisabi.Population)) {
	improveVPWMutation = func(fitness wabisabi.Fitness, selection wabisabi.Selection, population wabisabi.Population) (children wabisabi.Population) {
		victim := selection(population)
		victimGenome := victim.Genome.(Genotype)
		k := len(victimGenome)

		newGuyGenome := make(Genotype, k)
		copy(newGuyGenome, victimGenome)

		// fmt.Println("improveVPWMutation")
		// fmt.Println(victimGenome)
		// fmt.Println(victim.Meta["mvpw"].(float64))

		//randomly choose where to start looking for an item to trade
		startingPoint := rand.Intn(k)

		//modular scan over knapsack to find first item
		theRemovedItem := startingPoint
		done := false
		for i := mod(startingPoint+1, k); !done; i = mod(i+1, k) {
			if i == startingPoint {
				done = true
			} else {
				if newGuyGenome[i] {
					theRemovedItem = i
					done = true
				}
			}
		}
		//remove item
		// fmt.Println("removing", theRemovedItem)
		newGuyGenome[theRemovedItem] = false
		// fmt.Println(newGuyGenome)

		//calculate new knapsack weight
		newWeight := victim.Meta["weight"].(float64) - goodies[theRemovedItem].Weight
		remainingWeight := maxWeight - newWeight

		removedVPW := goodies[theRemovedItem].VPW

		//sample uniformly at random from the goodies which will now fit (beware sampling bias!)
		startingPoint = rand.Intn(k)
		theNewItem := startingPoint
		done = false
		for i := mod(startingPoint+1, k); !done; i = mod(i+1, k) {
			if i == startingPoint {
				done = true
			} else {
				//if we do not already have this item, and it has a better VPW than the one we removed, then take it
				if !newGuyGenome[i] && goodies[i].VPW > removedVPW && goodies[i].Weight <= remainingWeight {
					//take it
					theNewItem = i
					done = true
				}
			}
		}
		// fmt.Println("adding", theNewItem)

		//take the new item, perhaps it is the same thing we took out if nothing else fit
		newGuyGenome[theNewItem] = true
		// fmt.Println(newGuyGenome)
		newGuy := wabisabi.Individual{}
		newGuyScore, newGuyMeta := fitness(newGuyGenome)
		newGuy.Genome = newGuyGenome
		newGuy.Score = newGuyScore
		newGuy.Meta = newGuyMeta
		// fmt.Println(newGuyMeta["mvpw"].(float64))

		return wabisabi.Population{newGuy}
	}
	return
}

//remove an item, and replace it with an item of higher value
func makeImproveVMutation(goodies Goodies, maxWeight float64) (improveVMutation func(fitness wabisabi.Fitness, selection wabisabi.Selection, population wabisabi.Population) (children wabisabi.Population)) {
	improveVMutation = func(fitness wabisabi.Fitness, selection wabisabi.Selection, population wabisabi.Population) (children wabisabi.Population) {
		victim := selection(population)
		victimGenome := victim.Genome.(Genotype)
		k := len(victimGenome)

		newGuyGenome := make(Genotype, k)
		copy(newGuyGenome, victimGenome)

		// fmt.Println("improveVPWMutation")
		// fmt.Println(victimGenome)
		// fmt.Println(victim.Meta["mvpw"].(float64))

		//randomly choose where to start looking for an item to trade
		startingPoint := rand.Intn(k)

		//modular scan over knapsack to find first item
		theRemovedItem := startingPoint
		done := false
		for i := mod(startingPoint+1, k); !done; i = mod(i+1, k) {
			if i == startingPoint {
				done = true
			} else {
				if newGuyGenome[i] {
					theRemovedItem = i
					done = true
				}
			}
		}
		//remove item
		// fmt.Println("removing", theRemovedItem)
		newGuyGenome[theRemovedItem] = false
		// fmt.Println(newGuyGenome)

		//calculate new knapsack weight
		newWeight := victim.Meta["weight"].(float64) - goodies[theRemovedItem].Weight
		remainingWeight := maxWeight - newWeight

		removedV := goodies[theRemovedItem].Value

		//sample uniformly at random from the goodies which will now fit (beware sampling bias!)
		startingPoint = rand.Intn(k)
		theNewItem := startingPoint
		done = false
		for i := mod(startingPoint+1, k); !done; i = mod(i+1, k) {
			if i == startingPoint {
				done = true
			} else {
				//if we do not already have this item, and it has a better VPW than the one we removed, then take it
				if !newGuyGenome[i] && goodies[i].Value > removedV && goodies[i].Weight <= remainingWeight {
					//take it
					theNewItem = i
					done = true
				}
			}
		}
		// fmt.Println("adding", theNewItem)

		//take the new item, perhaps it is the same thing we took out if nothing else fit
		newGuyGenome[theNewItem] = true
		// fmt.Println(newGuyGenome)
		newGuy := wabisabi.Individual{}
		newGuyScore, newGuyMeta := fitness(newGuyGenome)
		newGuy.Genome = newGuyGenome
		newGuy.Score = newGuyScore
		newGuy.Meta = newGuyMeta
		// fmt.Println(newGuyMeta["mvpw"].(float64))

		return wabisabi.Population{newGuy}
	}
	return
}

func SwapMutation(fitness wabisabi.Fitness, selection wabisabi.Selection, population wabisabi.Population) (children wabisabi.Population) {
	victim := selection(population)
	victimGenome := victim.Genome.(Genotype)
	k := len(victimGenome)

	rightCut := 1 + rand.Intn(k-1)
	leftCut := rand.Intn(rightCut)

	//start with a clone of the victim genome
	childGenome := make(Genotype, k)
	copy(childGenome, victimGenome)

	temp := childGenome[leftCut]
	childGenome[leftCut] = childGenome[rightCut]
	childGenome[rightCut] = temp

	// fmt.Println()
	// fmt.Println("SWAP", leftCut, rightCut)
	// fmt.Println(victimGenome)
	// fmt.Println(childGenome)

	childScore, childMeta := fitness(childGenome)
	child := wabisabi.Individual{Score: childScore, Meta: childMeta, Genome: childGenome}
	children = wabisabi.Population{child}
	return
}

func SinglePointMutation(fitness wabisabi.Fitness, selection wabisabi.Selection, population wabisabi.Population) (children wabisabi.Population) {

	victim := selection(population)
	victimGenome := victim.Genome.(Genotype)
	k := len(victimGenome)
	//start with a clone of the victim genome
	childGenome := make(Genotype, k)
	copy(childGenome, victimGenome)

	//perform bit flip mutation at uniformly random position j
	j := rand.Intn(len(childGenome))
	childGenome[j] = !childGenome[j]

	// fmt.Println()
	// fmt.Println("SINGLE")
	// fmt.Println(victimGenome)
	// fmt.Println(childGenome)

	childScore, childMeta := fitness(childGenome)
	child := wabisabi.Individual{Score: childScore, Meta: childMeta, Genome: childGenome}
	children = wabisabi.Population{child}
	return
}

func MultiPointMutation(fitness wabisabi.Fitness, selection wabisabi.Selection, population wabisabi.Population) (children wabisabi.Population) {

	victim := selection(population)
	victimGenome := victim.Genome.(Genotype)
	k := len(victimGenome)

	//start with a clone of the victim genome
	childGenome := make(Genotype, k)
	copy(childGenome, victimGenome)
	//howMany := rand.Intn(k)
	howMany := 1 + int(math.Floor(math.Abs(rand.NormFloat64())))
	for i := 0; i < howMany; i++ {
		j := rand.Intn(k)
		childGenome[j] = !childGenome[j]
	}

	childScore, childMeta := fitness(childGenome)
	child := wabisabi.Individual{Score: childScore, Meta: childMeta, Genome: childGenome}
	children = wabisabi.Population{child}
	return
}

func MakeRandomPopulation(problemSize int, populationSize int, fitness wabisabi.Fitness) (population wabisabi.Population) {
	MakeRandomSolution := func() Genotype {
		s := make(Genotype, problemSize)
		for i := range s {
			if rand.Intn(2) == 1 {
				s[i] = true
			}
		}
		return s
	}

	population = make(wabisabi.Population, populationSize)
	for i := range population {
		tempSol := MakeRandomSolution()
		newGuy := wabisabi.Individual{Genome: tempSol}
		newGuy.Score, newGuy.Meta = fitness(tempSol)

		population[i] = newGuy
	}
	sort.Sort(population)
	return
}

func MakeFitness(goodies Goodies, maxWeight float64) (fitness wabisabi.Fitness) {
	// meanValuePerWeight := func(genome Genotype, goodies Goodies) (mvpw float64) {
	// 	x := 0.0
	// 	nItems := 0
	// 	for i := range genome {
	// 		if genome[i] {
	// 			x += goodies[i].VPW
	// 			nItems += 1
	// 		}
	// 	}
	// 	mvpw = math.Abs(x / float64(nItems))
	// 	return
	// }

	fitness = func(solution wabisabi.Genotype) (value float64, meta wabisabi.Meta) {
		knapsackSolution := solution.(Genotype)
		// score, value, weight, mvpw, meanValue := 0.0, 0.0, 0.0, 0.0, 0.0
		weight := 0.0
		itemCount := 0
		for i := range knapsackSolution {
			if knapsackSolution[i] {
				value += goodies[i].Value
				weight += goodies[i].Weight
				itemCount += 1
			}
		}

		if weight > maxWeight {
			value = -value
		}

		meta = make(wabisabi.Meta)
		meta["itemCount"] = itemCount
		// meta["value"] = value

		// if itemCount > 0 {
		// 	meta["meanValue"] = value / float64(itemCount)
		// 	meta["mvpw"] = meanValuePerWeight(knapsackSolution, goodies)
		// } else {
		// 	meta["meanValue"] = 0.0
		// 	meta["mvpw"] = 0.0
		// }

		meta["weight"] = weight

		//score = 0.4*value + 0.2*float64(itemCount) + 0.2*mvpw + 0.2*meanValue
		return
	}
	return
}

//this is a lovely idea which scales abysmally
func MemeticLocalSearch(fitness wabisabi.Fitness, selection wabisabi.Selection, population wabisabi.Population) wabisabi.Population {
	victim := selection(population)
	victimGenome := victim.Genome.(Genotype)
	k := len(victimGenome)
	//start with a clone of the victim genome
	childGenome := make(Genotype, k)
	copy(childGenome, victimGenome)
	best := victim
	for i := range victimGenome {
		newGenome := make(Genotype, k)
		copy(newGenome, victimGenome)
		newGenome[i] = !newGenome[i]
		child := wabisabi.Individual{Genome: newGenome}
		child.Score, child.Meta = fitness(newGenome)
		if child.Score > best.Score {
			best = child
		}
	}
	return wabisabi.Population{best}
}

func GreedyKnapsackSolver(goodies Goodies, maxWeight float64) (solution Genotype) {
	k := len(goodies)
	solution = make(Genotype, k)
	skipped := make(Genotype, k) //keep track of tasty things we could not fit

	weight := 0.0

	//take goodies in order of descending VPW until there is nothing more we can fit
	for weight <= maxWeight {
		tastiest := 0
		tastiestVPW := 0.0
		gotOne := false
		for i, goody := range goodies {
			//scan over goodies to find the tastiest one which we have not yet taken or skipped
			if !solution[i] && !skipped[i] && goody.VPW > tastiestVPW {
				tastiest = i
				tastiestVPW = goody.VPW
				gotOne = true
			}
		}

		//if we scanned the goodies, and did not find one, we are done
		if !gotOne {
			break
		}
		//found the tastiest, so take it if we can
		if weight+goodies[tastiest].Weight <= maxWeight {
			solution[tastiest] = true
			weight = weight + goodies[tastiest].Weight
		} else {
			skipped[tastiest] = true //we could not fit this one, don't try again later
		}
	}
	return
}

func main() {
	rand.Seed(42)

	problemSize := 1000
	maxWeight := 60.0

	goodies := MakeRandomGoodies(problemSize)
	fitness := MakeFitness(goodies, maxWeight)

	populationSize := 100000
	goalFitness := 1000000000.0
	maxGenerations := 100000
	eliteGroupSize := 1

	selection := utilities.MakeExponentialSelection(0.333)

	addItemMutation := makeAddItemMutation(goodies, maxWeight)
	removeItemMutation := makeRemoveItemMutation(goodies, maxWeight)
	improveVPWMutation := makeImproveVPWMutation(goodies, maxWeight)
	improveVMutation := makeImproveVMutation(goodies, maxWeight)
	splitItemMutation := makeSplitItemMutation(goodies, maxWeight)
	greedyAddMutation := makeGreedyAddMutation(goodies, maxWeight)
	operators := wabisabi.Operators{
		wabisabi.Operator{P: .82, Op: SinglePointCrossover},
		// wabisabi.Operator{P: .01, Op: MemeticLocalSearch},
		wabisabi.Operator{P: .05, Op: SinglePointMutation},
		wabisabi.Operator{P: .02, Op: addItemMutation},
		wabisabi.Operator{P: .02, Op: splitItemMutation},
		wabisabi.Operator{P: .02, Op: removeItemMutation},
		wabisabi.Operator{P: .02, Op: improveVPWMutation},
		wabisabi.Operator{P: .05, Op: greedyAddMutation},
		wabisabi.Operator{P: .02, Op: improveVMutation}}

	monitor := utilities.MakeSimpleMonitor(os.Stdout)

	initialPopulation := MakeRandomPopulation(problemSize, populationSize, fitness)

	initialState := wabisabi.Meta{"goalFitness": goalFitness, "maxGenerations": maxGenerations, "populationSize": populationSize, "eliteGroupSize": eliteGroupSize}
	_, finalPopulation := wabisabi.Evolve(fitness, selection, monitor, operators, initialPopulation, initialState)

	greedyGenome := GreedyKnapsackSolver(goodies, maxWeight)
	greedyScore, greedyMeta := fitness(greedyGenome)
	greedyGuy := wabisabi.Individual{Genome: greedyGenome, Score: greedyScore, Meta: greedyMeta}
	fmt.Println(greedyGuy)
	fmt.Println(finalPopulation[0])
}
