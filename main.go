package main

import (
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

// Config is the config-struct
type Config struct {
	Input          string
	Generations    int
	Initiate       string
	Estimator      string
	Mutate         string
	Dot            bool
	PopulationSize int
	MutationRate   float64
	Crossover      string
	Algorithmus    string
}

var cfg Config

func main() {

	readConfig(&cfg)
	// Seed the random gen
	rand.Seed(time.Now().UTC().UnixNano())

	// Extract the date out of the File and check if there could be errors
	verticesCount, customerDemand, Aij, Bij, Cij := parseFile(cfg.Input)
	costsA := inputToGraph(verticesCount, Aij)
	costsB := inputToGraph(verticesCount, Bij)
	costsC := inputToGraph(verticesCount, Cij)
	network, err := createNetwork(costsA, costsB, costsC)
	errFunc(err)

	// Generate a Graph based on the costs of A
	makeGraph(costsA)

	// Calculate the demand which is also the capacity of the source
	// The demand could be seen as negative storage capacity
	// This fact will be used later on
	// Also the last node is the source
	var sourceCapacity int64
	var demand []int64
	for i := range customerDemand {
		sourceCapacity = sourceCapacity + customerDemand[i]
		demand = append(demand, -1*customerDemand[i])
	}
	demand = append(demand, sourceCapacity)

	if cfg.Algorithmus == "hillclimber" {
		for i := 0; i < 200; i++ {
			println("------ ", i, ". Run ------")
			hillclimb(cfg, verticesCount, demand, network, costsA, costsB, costsC)
		}
	} else if cfg.Algorithmus == "evolutionär" {
		println("Generating population...")
		var population []Child
		population = populate(network, verticesCount, demand, costsA, costsB, costsC)
		println("Population generated...")
		for i := 0; i < cfg.Generations; i++ {
			population = ranking(population)
			printChild(population[0], i)
			population = selection(population, costsA, costsB, costsC)
		}
	}
}

func printChild(x Child, n int) {
	print(n, ",")
	for i := range x.storage {
		print(x.storage[i], ",")
	}
	println(x.fitness, ",")
}

func hillclimb(cfg Config, verticesCount int, demand []int64, network [][]bool, costsA, costsB, costsC [][]int64) {

	c := new(Child) // Always Child
	x := new(Child) // Always parent
	x.demand = make([]int64, verticesCount)
	copied := copy(x.demand, demand)

	print(copied, " - ")
	for i := range demand {
		print(demand[i], ", ")
	}
	println()

	// Initiate the flow but make it dependend from the config
	if cfg.Initiate == "zero" {
		x.initiateFlowZero(verticesCount)
	} else if cfg.Initiate == "one" {
		x.initiateFlowOne(verticesCount, network)
	} else if cfg.Initiate == "two" {
		x.initiateFlowTwo(verticesCount, network)
	}

	print(len(x.storage), " - ")
	for _, storage := range x.storage {
		print(storage, ", ")
	}
	x.costCalculator(costsA, costsB, costsC)
	println(x.fitness, ",")

	for i := 0; i < cfg.Generations; i++ {
		x.findNeighbourTwo(c, network)
		c.costCalculator(costsA, costsB, costsC)
		if c.fitness < x.fitness {
			c.toParent(x)
			print(len(x.storage), " - ")
			for k := range x.storage {
				print(x.storage[k], ", ")
			}
			println(x.fitness, ",")
		}
	}
	print(len(demand), " - ")
	for i := range demand {
		print(demand[i], ", ")
	}
	println()
}

func readConfig(cfg *Config) {
	f, err := os.Open("cfg.yml")
	errFunc(err)
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(cfg)
	errFunc(err)
}

func makeGraph(network [][]int64) {
	graph := `digraph graphname
	{
`
	for i, row := range network {
		for j := range row {
			if network[i][j] != 0 {
				graph = graph + "    " + strconv.Itoa(i) + " -> " + strconv.Itoa(j) + "[ label=" + strconv.FormatInt(network[i][j], 10) + "];\n"
			}
		}
	}
	graph = graph + "}"

	// Dump dot-graph to file
	f, err := os.Create("input.dot")
	errFunc(err)
	_, err = f.WriteString(graph)
	errFunc(err)
}

func errFunc(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
