package controller

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gocarina/gocsv"
)

// type Node struct {
// 	ID           int                `json:"id"`
// 	Humidity     float64            `json:"humidity"`
// 	Temperature  float64            `json:"temperature"`
// 	StepCount    float64            `json:"step_count"`
// 	StressLevel  float64            `json:"-"`
// 	Result       map[string]float64 `json:"-"`
// 	ChosenResult float64            `json:"-"`
// 	ClusterCode  string             `json:"-"`
// }

func UploadDataset(rw http.ResponseWriter, r *http.Request) {
	// Parse our multipart form, 10 << 20 specifies a maximum
	// upload of 10 MB files.
	r.ParseMultipartForm(10 << 20)
	// FormFile returns the first file for the given key `myFile`
	// it also returns the FileHeader so we can get the Filename,
	// the Header and the size of the file
	file, handler, err := r.FormFile("datasetFile")
	if err != nil {
		log.Printf("gagal dapatkan datasetFile")
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte("Harap upload file dataset\n"))
		return

	}
	defer file.Close()
	fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	fmt.Printf("File Size: %+v\n", handler.Size)
	fmt.Printf("MIME Header: %+v\n", handler.Header)

	// csvLines, readErr := csv.NewReader(file).ReadAll()
	// if readErr != nil {
	// 	//handle error
	// }

	tempFile, err := ioutil.TempFile("files", "dataset-*.csv")
	if err != nil {
		log.Printf("Gagal membaca data %s", err.Error())
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte("Gagal membaca dataset\n"))
		return
	}

	filename := tempFile.Name()

	defer func(f string) {
		tempFile.Close()

		e := os.Remove(f)
		if e != nil {
			log.Printf("Gagal menghapus file %s", err.Error())
		}
	}(filename)

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Printf("Gagal menyimpan data %s", err.Error())
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte("Gagal menyimpan dataset\n"))
		return
	}
	// write this byte array to our temporary file
	tempFile.Write(fileBytes)

	if err := saveData(filename); err != nil {
		log.Printf("Gagal memproses data %s", err.Error())
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte("Gagal memproses dataset\n"))
		return
	}

	Write(rw, http.StatusOK, "berhasil simpan data")
}

var (
	mainNodes    []Node
	nodeKMeans   []Node
	nodeKMedoids []Node
)

func saveData(filename string) error {
	fmt.Println("filename: ", filename)
	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := gocsv.UnmarshalFile(f, &mainNodes); err != nil { // Load clients from file
		return err
	}

	for k := range mainNodes {
		//fmt.Println("node: ", mainNodes[k])
		mainNodes[k].ID = k + 1
	}

	nodeKMeans = mainNodes
	nodeKMedoids = mainNodes

	return nil

}
