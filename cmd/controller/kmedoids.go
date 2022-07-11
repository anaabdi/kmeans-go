package controller

import (
	"encoding/csv"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
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
	TotalCluster              int                `json:"total_cluster"`
	TotalIteration            int                `json:"total_iteration"`
	Duration                  string             `json:"duration"`
	InitialCentroid           map[string]string  `json:"initial_centroids"`
	Results                   map[string]string  `json:"results"`
	StandarDeviationOfCluster map[string]float64 `json:"standard_deviation_clusters"`
	HighestStandardDeviation  float64            `json:"highest_standard_deviation"`
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

	// if len(req.InitialCentroidRowIDs) != req.KExact {
	// 	log.Printf("banyak initial centroids tidak sama dengan nilai K")
	// 	rw.WriteHeader(http.StatusBadRequest)
	// 	rw.Write([]byte("banyak initial centroids tidak sama dengan nilai K\n"))
	// 	return
	// }

	k := req.KExact
	fmt.Println("K: ", k)

	mainNodes := nodeKMedoids

	initialCentroids := make(map[string]string, k)

	initRowIDCentroid := []int{}
	if len(req.InitialCentroidRowIDs) == 0 {
		initRowIDCentroid = getInitialCentroids(k)
	} else {
		initRowIDCentroid = req.InitialCentroidRowIDs
	}

	fmt.Println("initial row id chosen as centroid: ", initRowIDCentroid)

	mapOfCentroid := make(map[string]Node, k)
	for i, rowID := range initRowIDCentroid {
		id := fmt.Sprintf("C%d", i+1)

		for _, node := range mainNodes {
			if node.ID == rowID {
				initialCentroids[id] = fmt.Sprintf("%.2f, %.2f, %.2f",
					node.Humidity, node.Temperature, node.StepCount)

				mapOfCentroid[id] = node
				break
			}
		}
	}

	if len(mapOfCentroid) != len(initRowIDCentroid) {
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

		if iteration == 1 {
			fmt.Println("INITIAL CENTROIDS: ", mapOfCentroid)
		}

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

			// fmt.Printf("result chosen: %.2f %.2f %.2f %s %.2f \n", mainNodes[k].Humidity,
			// 	mainNodes[k].Result["C1"], mainNodes[k].Result["C2"],
			// 	chosenClusterCode, mainNodes[k].ChosenResult)

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
			fmt.Println("Finding New Centroid next iteration: ", iteration+1)
			newCentroidNodeIDs := map[int]bool{}

			for clusterNode, node := range mapOfCentroid {
				newCentroid := getNewCentroidKMedoids(mainNodes, node.ID, newCentroidNodeIDs)
				fmt.Printf("New Centroid ITERATION %d: %s: %v \n", iteration+1, clusterNode, newCentroid)
				mapOfCentroid[clusterNode] = newCentroid

				newCentroidNodeIDs[newCentroid.ID-1] = true
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
		TotalCluster:              k,
		Duration:                  fmt.Sprintf("%fs", time.Since(timeStart).Seconds()),
		Results:                   make(map[string]string, len(chosenMapOfCluterResults)),
		InitialCentroid:           initialCentroids,
		TotalIteration:            iteration,
		StandarDeviationOfCluster: make(map[string]float64, len(chosenMapOfCluterResults)),
	}

	var highestStandardDev float64

	for k, nodes := range chosenMapOfCluterResults {
		var node []string
		var data []float64
		for _, v := range nodes {
			node = append(node, strconv.Itoa(v.ID))

			data = append(data, float64(v.ID))
		}

		fmt.Println("Cluster: ", k, node)
		resp.Results[k] = strings.Join(node, ", ")

		standarDeviation := calcStandardDeviation(nodes)
		//standarDeviation, _ := stats.StandardDeviationSample(data)
		resp.StandarDeviationOfCluster[k] = math.Floor(standarDeviation*100) / 100

		if highestStandardDev < standarDeviation {
			highestStandardDev = standarDeviation
		}

		go func(nodes []Node, cluster string) {
			saveClusterToCSV("kmedoids", cluster, req.KExact, nodes)
		}(nodes, k)
	}

	resp.HighestStandardDeviation = math.Floor(highestStandardDev*100) / 100

	file, err := os.Create(fmt.Sprintf("files/%s%d-summary.csv", "kmedoids", req.KExact))
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// write header
	writer.Write([]string{"Total Iterasi", strconv.Itoa(resp.TotalIteration)})
	writer.Write([]string{"Total Durasi", resp.Duration})

	for c, v := range resp.InitialCentroid {
		writer.Write([]string{fmt.Sprintf("Centroid Awal %s", c), v})
	}

	for k, v := range resp.StandarDeviationOfCluster {
		fmt.Printf("Standar Deviation of Cluster %s : %.2f \n", k, v)

		writer.Write([]string{fmt.Sprintf("Standar Deviasi %s", k), fmt.Sprintf("%.2f", v)})
	}

	for k := range resp.StandarDeviationOfCluster {
		writer.Write([]string{fmt.Sprintf("Banyak Anggota %s", k), fmt.Sprintf("%d", len(mapOfClusterResults[k]))})
	}

	writer.Write([]string{"Standar Deviasi Tertinggi", fmt.Sprintf("%.2f", resp.HighestStandardDeviation)})

	fmt.Printf("Highest Standar Deviation of All Cluster: %.2f \n", resp.HighestStandardDeviation)

	Write(rw, http.StatusOK, resp)

}

func getRandomNumber(min, max int, otherCentroidNodeIDs map[int]bool) int {
	rand.Seed(time.Now().UnixNano())
	k := rand.Intn(max-min) + min

	// fmt.Println("otherCentroidNodeIDs: ", otherCentroidNodeIDs)
	// fmt.Println("random node id: ", k)

	if otherCentroidNodeIDs[k] {
		k = getRandomNumber(min, max, otherCentroidNodeIDs)
	}

	return k
}

func getNewCentroidKMedoids(nodes []Node, currentNodeID int, otherCentroidNodeIDs map[int]bool) Node {
	otherCentroidNodeIDs[currentNodeID-1] = true

	k := getRandomNumber(0, len(nodes), otherCentroidNodeIDs)
	return nodes[k]
}
