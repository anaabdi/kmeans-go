package controller

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type KMedoidsRequest struct {
	KExact   int `json:"k_exact"`
	RangeMin int `json:"range_min"`
	RangeMax int `json:"range_max"`
}

type KMedoidsResponse struct {
	TotalCluster    int               `json:"total_cluster"`
	TotalIteration  int               `json:"total_iteration"`
	Duration        string            `json:"duration"`
	InitialCentroid map[string]string `json:"initial_centroids"`
	Results         map[string]string `json:"results"`
}

func KMedoidsController(rw http.ResponseWriter, r *http.Request) {
	log.Printf("Hello from %s", r.RemoteAddr)

	var req KMeansRequest
	r.ParseForm()
	req.RangeMin, _ = strconv.Atoi(r.FormValue("range_min"))
	req.RangeMax, _ = strconv.Atoi(r.FormValue("range_max"))
	req.KExact, _ = strconv.Atoi(r.FormValue("k_exact"))
	req.InitialCentroidRowIDs = getRowIDsFromString(r.FormValue("initial_centroid_row_ids"))

	if req.KExact < 2 || req.KExact > 9 {
		log.Printf("Harap sertakan range min dan max nya")
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte("Harap sertakan range min dan max nya\n"))
		return
	}

	if len(req.InitialCentroidRowIDs) != req.KExact {
		log.Printf("banyak initial centroids tidak sama dengan nilai K")
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte("banyak initial centroids tidak sama dengan nilai K\n"))
		return
	}

	k := req.KExact
	fmt.Println("K: ", k)

	//c1, indexChosen := getInitialRandomCentroid(mainNodes)

	//c1, c2 := getInitialCentroid(mainNodes)

	initialCentroids := make(map[string]string, k)

	mapOfCentroid := make(map[string]Node, k)
	for i, rowID := range req.InitialCentroidRowIDs {
		id := fmt.Sprintf("C%d", i+1)
		// if i == 0 {
		// 	mapOfCentroid[id] = c1
		// 	initialCentroids[id] = fmt.Sprintf("%.2f, %.2f, %.2f", c1.Humidity, c1.Temperature, c1.StepCount)
		// } else {
		// 	c := mainNodes[indexChosen+1]
		// 	initialCentroids[id] = fmt.Sprintf("%.2f, %.2f, %.2f", c.Humidity, c.Temperature, c.StepCount)
		// 	mapOfCentroid[id] = c
		// }

		for _, node := range mainNodes {
			if node.ID == rowID {
				initialCentroids[id] = fmt.Sprintf("%.2f, %.2f, %.2f",
					node.Humidity, node.Temperature, node.StepCount)

				mapOfCentroid[id] = node
				break
			}
		}

		// if i == 0 {
		// 	mapOfCentroid[id] = c1
		// 	initialCentroids[id] = fmt.Sprintf("%.2f, %.2f, %.2f", c1.Humidity, c1.Temperature, c1.StepCount)
		// } else {
		// 	mapOfCentroid[id] = c2
		// 	initialCentroids[id] = fmt.Sprintf("%.2f, %.2f, %.2f", c2.Humidity, c2.Temperature, c2.StepCount)
		// }
	}

	if len(mapOfCentroid) != len(req.InitialCentroidRowIDs) {
		log.Printf("Ada row id tidak ditemukan sebagai pilihan centroid awal")
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte("Ada row id tidak ditemukan sebagai pilihan centroid awal\n"))
		return
	}

	var mapOfClusterResults map[string][]Node

	var storedIterationResult []map[string][]Node

	var previousTotalResult float64

	timeStart := time.Now()
	iteration := 0
	for {
		iteration++

		fmt.Println("start iteration: ", iteration)

		fmt.Println("CENTROIDS: ", mapOfCentroid)
		mapOfClusterResults = make(map[string][]Node, 0) // C1 / C2 / C3

		var totalResult float64

		for k, v := range mainNodes {

			mainNodes[k].Result = make(map[string]float64, len(mapOfCentroid))

			for centroidKey, centroid := range mapOfCentroid {
				humidity := math.Pow(v.Humidity-centroid.Humidity, 2)
				temperature := math.Pow(v.Temperature-centroid.Temperature, 2)
				stepCount := math.Pow(v.StepCount-centroid.StepCount, 2)

				result := math.Sqrt(humidity + temperature + stepCount)

				//fmt.Println("centroid key result: ", k, centroidKey, result)

				mainNodes[k].Result[centroidKey] = math.Round(result*100) / 100

			}

			chosenClusterCode := defineTheCluster(mainNodes[k].Result)

			mainNodes[k].ClusterCode = chosenClusterCode // C1 / C2 / C3
			mainNodes[k].ChosenResult = mainNodes[k].Result[chosenClusterCode]

			mapOfClusterResults[chosenClusterCode] = append(
				mapOfClusterResults[chosenClusterCode], mainNodes[k])

			fmt.Printf("result chosen: %.2f %.2f %.2f %s %.2f \n", mainNodes[k].Humidity,
				mainNodes[k].Result["C1"], mainNodes[k].Result["C2"],
				chosenClusterCode, mainNodes[k].ChosenResult)

			totalResult += mainNodes[k].ChosenResult
		}

		storedIterationResult = append(storedIterationResult, mapOfClusterResults)

		if previousTotalResult == 0 {
			previousTotalResult = totalResult
		}

		fmt.Printf("result: %.2f - %.2f \n", totalResult, previousTotalResult)
		subResult := totalResult - previousTotalResult
		fmt.Println("end of iteration: ", iteration)
		if subResult <= 0 {
			// Pilih centroid
			// newCentroid1, newCentroid2 := getSecondCentroid(mainNodes)
			// mapOfCentroid["C1"] = newCentroid1
			// mapOfCentroid["C2"] = newCentroid2

			for clusterNode, node := range mapOfCentroid {
				mapOfCentroid[clusterNode] = getNewCentroidKMedoids(mainNodes, node.ID)
			}

			previousTotalResult = totalResult

		} else {
			fmt.Println("stop iteration: ", iteration)
			break
		}
	}

	//timeStop := time.Now()

	fmt.Println("duration: ", time.Since(timeStart))

	fmt.Println("PRINTING RESULTs")

	chosenIteration := iteration - 2 // - 1 = current / -2 == previous
	fmt.Println("chosen iteration: ", chosenIteration)
	chosenMapOfCluterResults := storedIterationResult[chosenIteration]

	resp := KMeansResponse{
		TotalCluster:    k,
		Duration:        fmt.Sprintf("%fs", time.Since(timeStart).Seconds()),
		Results:         make(map[string]string, len(chosenMapOfCluterResults)),
		InitialCentroid: initialCentroids,
		TotalIteration:  iteration,
	}

	for k, nodes := range chosenMapOfCluterResults {
		var node []string
		for _, v := range nodes {
			node = append(node, strconv.Itoa(v.ID))
		}

		fmt.Println("Cluster: ", k, node)
		resp.Results[k] = strings.Join(node, ", ")
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

	Write(rw, http.StatusOK, resp)

}

func getRandomNumber(min, max, avoidNumber int) int {
	rand.Seed(time.Now().UnixNano())
	k := rand.Intn(max-min) + min

	if k == avoidNumber {
		k = getRandomNumber(min, max, avoidNumber)
	}

	return k
}

func getNewCentroidKMedoids(nodes []Node, currentNodeID int) Node {
	k := getRandomNumber(0, len(nodes), currentNodeID)
	return nodes[k]
}
