package main

import (
	"log"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

// Config is the config-struct
type Config struct {
	Input       string
	Generations int
	Initiate    string
	Estimator   string
	Mutate      string
	Dot         bool
}

func main() {
	cfg := Config{}
	readConfig(&cfg)

	var c Child // Always Child
	var x Child // Always parent
	maxGenerations := cfg.Generations

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
	var sourceCapacity int64
	for _, demand := range customerDemand {
		sourceCapacity = sourceCapacity + demand
	}

	if cfg.Initiate == "zero" {
		x.initiateFlowZero(verticesCount)
	} else if cfg.Initiate == "one" {
		x.initiateFlowOne(verticesCount, customerDemand, network)
	} else if cfg.Initiate == "two" {
		x.initiateFlowTwo(verticesCount, customerDemand, network, sourceCapacity)
	}

	x.costCalculator(costsA, costsB, costsC, customerDemand)
	println(x.fitness)

	for i := 0; i < maxGenerations; i++ {
		x.findNeighbour(&c)
		c.costCalculator(costsA, costsB, costsC, customerDemand)
		if c.fitness < x.fitness {
			c.toParent(&x)
			println(c.fitness)
		}
	}

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
