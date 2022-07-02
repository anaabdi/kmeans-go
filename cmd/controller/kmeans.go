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
	TotalCluster              int                `json:"total_cluster"`
	TotalIteration            int                `json:"total_iteration"`
	Duration                  string             `json:"duration"`
	InitialCentroid           map[string]string  `json:"initial_centroids"`
	Results                   map[string]string  `json:"results"`
	StandarDeviationOfCluster map[string]float64 `json:"standard_deviation_clusters"`
	HighestStandardDeviation  float64            `json:"highest_standard_deviation"`
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

	// if len(req.InitialCentroidRowIDs) != req.KExact {
	// 	log.Printf("banyak initial centroids tidak sama dengan nilai K")
	// 	rw.WriteHeader(http.StatusBadRequest)
	// 	rw.Write([]byte("banyak initial centroids tidak sama dengan nilai K\n"))
	// 	return
	// }

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

	mainNodes := nodeKMeans

	// c1, indexChosen := getInitialRandomCentroid(mainNodes)

	//c1, c2 := getInitialCentroid(mainNodes)

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
				log.Println("step count: ", node.StepCount)
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
		TotalCluster:              k,
		Duration:                  fmt.Sprintf("%fs", time.Since(timeStart).Seconds()),
		Results:                   make(map[string]string, len(mapOfClusterResults)),
		InitialCentroid:           initialCentroids,
		TotalIteration:            iteration,
		StandarDeviationOfCluster: make(map[string]float64, len(mapOfClusterResults)),
	}

	var highestStandardDev float64

	for k, nodes := range mapOfClusterResults {
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
	}

	resp.HighestStandardDeviation = math.Floor(highestStandardDev*100) / 100

	for k, v := range resp.StandarDeviationOfCluster {
		fmt.Printf("Standar Deviation of Cluster %s : %.2f \n", k, v)
	}

	fmt.Printf("Highest Standar Deviation of All Cluster: %.2f \n", resp.HighestStandardDeviation)

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

func calcStandardDeviation(nodes []Node) float64 {
	var sum, mean, sd float64
	for _, v := range nodes {
		sum += float64(v.ID)
	}

	totalNodes := float64(len(nodes))

	mean = sum / totalNodes

	for _, v := range nodes {
		sd += math.Pow(float64(v.ID)-mean, 2)
	}

	// The use of Sqrt math function func Sqrt(x float64) float64
	sd = math.Sqrt(sd / (totalNodes - 1))

	fmt.Println("The Standard Deviation is : ", sd)

	return math.Floor(sd*100) / 100
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

	fmt.Printf("results defining the cluster: %v: chosen %s \n", result, lowestValueKey)
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
