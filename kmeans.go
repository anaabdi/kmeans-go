package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/gocarina/gocsv"
)

type Node struct {
	ID          int
	Humidity    float64
	Temperature float64
	StepCount   float64
	StressLevel float64
	Result      map[string]float64
	ClusterCode string
}

func main() {
	f, err := os.OpenFile("Stress-Lysis.csv", os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	var mainNodes []Node

	if err := gocsv.UnmarshalFile(f, &mainNodes); err != nil { // Load clients from file
		panic(err)
	}

	for k := range mainNodes {
		mainNodes[k].ID = k + 1
	}

	min := 2
	max := 4
	rand.Seed(time.Now().UnixNano())
	k := rand.Intn(max-min) + min
	fmt.Println("K: ", k)

	c1, indexChosen := getInitialRandomCentroid(mainNodes)

	mapOfCentroid := make(map[string]Node, k)
	for i := 0; i < k; i++ {
		id := fmt.Sprintf("C%d", i+1)
		if i == 0 {
			mapOfCentroid[id] = c1
		} else {
			mapOfCentroid[id] = mainNodes[indexChosen+1]
		}
	}

	// for _, v := range mapOfCentroid {
	// 	fmt.Println("centroid: ", v)
	// }

	// return

	var mapOfClusterResults map[string][]Node

	timeStart := time.Now()
	interation := 0
	for {
		interation++

		fmt.Println("start iteration: ", interation)

		fmt.Println("CENTROIDS: ", mapOfCentroid)
		mapOfClusterResults = make(map[string][]Node, 0) // C1 / C2 / C3

		for k, v := range mainNodes {
			for centroidKey, centroid := range mapOfCentroid {
				humidity := math.Pow(v.Humidity-centroid.Humidity, 2)
				temperature := math.Pow(v.Temperature-centroid.Temperature, 2)
				stepCount := math.Pow(v.StepCount-centroid.StepCount, 2)

				result := math.Sqrt(humidity + temperature + stepCount)

				//fmt.Println("centroid key result: ", k, centroidKey, result)

				mainNodes[k].Result = map[string]float64{
					centroidKey: math.Round(result*100) / 100,
				}
			}

			chosenClusterCode := defineTheCluster(mainNodes[k].Result)

			mainNodes[k].ClusterCode = chosenClusterCode // C1 / C2 / C3

			mapOfClusterResults[chosenClusterCode] = append(
				mapOfClusterResults[chosenClusterCode], mainNodes[k])

		}

		var mustStopIteration = true
		for clusterNode, results := range mapOfClusterResults {
			currentCentroid := mapOfCentroid[clusterNode]
			newCentroid := getNewCentroid(results)

			mapOfCentroid[clusterNode] = newCentroid
			if mustStopIteration && isCentroidSame(currentCentroid, newCentroid) {
				mustStopIteration = true
			} else {
				mustStopIteration = false
			}
		}

		fmt.Println("end of iteration: ", interation)
		if mustStopIteration {
			fmt.Println("stop iteration: ", interation)
			break
		}
	}

	//timeStop := time.Now()

	fmt.Println("duration: ", time.Since(timeStart))

	fmt.Println("PRINTING RESULTs")
	for k, nodes := range mapOfClusterResults {
		var node []int
		for _, v := range nodes {
			node = append(node, v.ID)
		}

		fmt.Println("Cluster: ", k, node)
	}

	/*
		# k = 2
		# pilih centroid secara acak, C1 dan C2
		# cek jarak masing2 data dari row 1 ke C1 dan C2
		# sqroot (h1-c1)2 + (t1-c1)2 + (s1-c1)2
		# sqroot (h1-c2)2 + (t1-c2)2 + (s1-c2)2
		# hasil nya ambil yang kecil, C1 atau C2 dari masing2 row
		# cari nilai centroid baru
		# C1 = average of sum
		# C2 = average of sum
	*/
}

func isCentroidSame(current, new Node) bool {
	return current.Humidity == new.Humidity && current.Temperature == new.Temperature && current.StepCount == new.StepCount
}

func defineTheCluster(result map[string]float64) string {
	// returning chosen cluster id
	min := result["C1"]
	lowestValueKey := "C1"
	for k := range result {
		if min == 0.0 {
			//fmt.Printf("result %s %f %f \n", k, result[k], min)
			min = result[k]
			lowestValueKey = k
			continue
		}

		if result[k] < min {
			//fmt.Printf("result replace %s %f %f \n", k, result[k], min)
			min = result[k]
			lowestValueKey = k
		}
	}
	return lowestValueKey
}

func getInitialRandomCentroid(nodes []Node) (Node, int) {
	rand.Seed(time.Now().UnixNano())
	totalNode := len(nodes)
	randId := rand.Intn((totalNode/2)-2+1) + 2
	return nodes[randId], randId
}

func getNewCentroid(nodes []Node) Node {
	humiditySum := 0.0
	tempSum := 0.0
	stepSum := 0.0
	totalData := float64(len(nodes))

	for _, v := range nodes {
		humiditySum += v.Humidity
		tempSum += v.Humidity
		stepSum += v.Humidity
	}

	humidityAvg := humiditySum / totalData
	stepAvg := stepSum / totalData
	tempAvg := tempSum / totalData

	return Node{
		Humidity:    math.Round(humidityAvg*100) / 100,
		Temperature: math.Round(tempAvg*100) / 100,
		StepCount:   math.Round(stepAvg*100) / 100,
	}
}

func median(data []Node) *Node {

	var median *Node
	l := len(data)
	if l == 0 {
		return nil
	} else if l%2 == 0 {
		median = &data[l/2-1]
	} else {
		median = &data[l/2]
	}

	return median
}
