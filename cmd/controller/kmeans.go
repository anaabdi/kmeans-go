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

type Node struct {
	ID           int                `json:"id" csv:"-"`
	Humidity     float64            `json:"humidity" csv:"Humidity"`
	Temperature  float64            `json:"temperature" csv:"Temperature"`
	StepCount    float64            `json:"step_count" csv:"Step count"`
	StressLevel  float64            `json:"-" csv:"-"`
	Result       map[string]float64 `json:"-" csv:"-"`
	ChosenResult float64            `json:"-" csv:"-"`
	ClusterCode  string             `json:"-" csv:"-"`
}

type KMeansRequest struct {
	KExact                int   `json:"k_exact"`
	RangeMin              int   `json:"range_min"`
	RangeMax              int   `json:"range_max"`
	InitialCentroidRowIDs []int `json:"initial_centroid_row_ids"`
}

type KMeansResponse struct {
	TotalCluster    int               `json:"total_cluster"`
	TotalIteration  int               `json:"total_iteration"`
	Duration        string            `json:"duration"`
	InitialCentroid map[string]string `json:"initial_centroids"`
	Results         map[string]string `json:"results"`
}

func getRowIDsFromString(n string) []int {
	ids := strings.Split(n, ",")
	var result []int
	for _, id := range ids {
		id = strings.TrimSpace(id)
		idInteger, err := strconv.Atoi(id)
		if err != nil {
			log.Printf("ID Row nya tidak valid %s", id)
			continue
		}
		result = append(result, idInteger)
	}

	return result
}

func KMeansController(rw http.ResponseWriter, r *http.Request) {
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

	log.Println("length of data: ", len(mainNodes))

	// f, err := os.OpenFile("Stress-Lysis.csv", os.O_RDWR|os.O_CREATE, os.ModePerm)
	// if err != nil {
	// 	panic(err)
	// }
	// defer f.Close()

	// var mainNodes []Node

	// if err := gocsv.UnmarshalFile(f, &mainNodes); err != nil { // Load clients from file
	// 	panic(err)
	// }

	// for k := range mainNodes {
	// 	//fmt.Println("node: ", mainNodes[k])
	// 	mainNodes[k].ID = k + 1
	// }

	k := req.KExact
	fmt.Println("K: ", k)

	// c1, indexChosen := getInitialRandomCentroid(mainNodes)

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
				log.Println("step count: ", node.StepCount)
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

	timeStart := time.Now()
	iteration := 0
	for {
		iteration++

		fmt.Println("start iteration: ", iteration)

		fmt.Println("CENTROIDS: ", mapOfCentroid)
		mapOfClusterResults = make(map[string][]Node, 0) // C1 / C2 / C3

		for k, v := range mainNodes {
			mainNodes[k].Result = make(map[string]float64, len(mapOfCentroid))

			for centroidKey, centroid := range mapOfCentroid {
				humidity := math.Pow(v.Humidity-centroid.Humidity, 2)
				temperature := math.Pow(v.Temperature-centroid.Temperature, 2)
				stepCount := math.Pow(v.StepCount-centroid.StepCount, 2)

				result := math.Sqrt(humidity + temperature + stepCount)

				mainNodes[k].Result[centroidKey] = math.Round(result*100) / 100
			}

			chosenClusterCode := defineTheCluster(mainNodes[k].Result)

			mainNodes[k].ClusterCode = chosenClusterCode // C1 / C2 / C3

			// fmt.Printf("%s %.2f - %.2f %.2f \n", mainNodes[k].ClusterCode,
			// 	mainNodes[k].Humidity, mainNodes[k].Result["C1"], mainNodes[k].Result["C2"])

			mapOfClusterResults[chosenClusterCode] = append(
				mapOfClusterResults[chosenClusterCode], mainNodes[k])

		}

		var mustStopIteration = true
		for clusterNode, results := range mapOfClusterResults {
			currentCentroid := mapOfCentroid[clusterNode]
			newCentroid := getNewCentroidKMeans(results)

			mapOfCentroid[clusterNode] = newCentroid
			if mustStopIteration && isCentroidSame(currentCentroid, newCentroid) {
				mustStopIteration = true
			} else {
				mustStopIteration = false
			}
		}

		fmt.Println("end of iteration: ", iteration)
		if mustStopIteration {
			fmt.Println("stop iteration: ", iteration)
			break
		}
	}

	fmt.Println("duration: ", time.Since(timeStart))

	fmt.Println("PRINTING RESULTs")

	resp := KMeansResponse{
		TotalCluster:    k,
		Duration:        fmt.Sprintf("%fs", time.Since(timeStart).Seconds()),
		Results:         make(map[string]string, len(mapOfClusterResults)),
		InitialCentroid: initialCentroids,
		TotalIteration:  iteration,
	}

	for k, nodes := range mapOfClusterResults {
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

func isCentroidSame(current, new Node) bool {
	return current.Humidity == new.Humidity && current.Temperature == new.Temperature && current.StepCount == new.StepCount
}

func defineTheCluster(result map[string]float64) string {
	// returning chosen cluster id
	min := result["C1"]
	lowestValueKey := "C1"
	for k := range result {
		if result[k] < min {
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

func getInitialCentroid(nodes []Node) (Node, Node) {
	var c1 Node
	var c2 Node
	for _, v := range nodes {
		if v.Humidity == 18.16 {
			c1 = v
		}

		if v.Humidity == 23.61 {
			c2 = v
		}
	}

	return c1, c2
}

func getSecondCentroid(nodes []Node) (Node, Node) {
	var c1 Node
	var c2 Node
	for _, v := range nodes {
		if v.Humidity == 27.12 {
			c1 = v
		}

		if v.Humidity == 27.64 {
			c2 = v
		}
	}

	return c1, c2
}

func getNewCentroidKMeans(nodes []Node) Node {
	humiditySum := 0.0
	tempSum := 0.0
	stepSum := 0.0
	totalData := float64(len(nodes))

	for _, v := range nodes {
		humiditySum += v.Humidity
		tempSum += v.Temperature
		stepSum += v.StepCount
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
