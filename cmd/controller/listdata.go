package controller

import (
	"fmt"
	"net/http"
)

var (
	sharedInitialCentroids []int
)

func getInitialCentroids(kExact int) []int {
	if sharedInitialCentroids == nil {

		sharedInitialCentroids = make([]int, 0)

		otherCentroidNodeIDs := make(map[int]bool, 0)

		for i := 0; i < kExact; i++ {
			k := getRandomNumber(0, len(mainNodes), otherCentroidNodeIDs)

			otherCentroidNodeIDs[k+1] = true

			sharedInitialCentroids = append(sharedInitialCentroids, k+1)
		}

		fmt.Println("sharedInitialCentroids new: ", sharedInitialCentroids)

		return sharedInitialCentroids
	}

	if len(sharedInitialCentroids) != kExact {
		// Reset

		sharedInitialCentroids = make([]int, 0)

		otherCentroidNodeIDs := make(map[int]bool, 0)

		for i := 0; i < kExact; i++ {
			k := getRandomNumber(0, len(mainNodes), otherCentroidNodeIDs)

			otherCentroidNodeIDs[k+1] = true

			sharedInitialCentroids = append(sharedInitialCentroids, k+1)
		}

		fmt.Println("sharedInitialCentroids new: ", sharedInitialCentroids)

		return sharedInitialCentroids
	}

	fmt.Println("sharedInitialCentroids existing: ", sharedInitialCentroids)
	return sharedInitialCentroids
}

func ListDataController(rw http.ResponseWriter, r *http.Request) {
	if len(mainNodes) == 0 {
		Write(rw, http.StatusOK, map[string][]Node{"datasets": []Node{}})
		return
	}
	Write(rw, http.StatusOK, map[string][]Node{"datasets": mainNodes})
}
