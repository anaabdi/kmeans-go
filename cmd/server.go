package main

import (
	"context"
	"fmt"

	"machinelearning/cmd/controller"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/gocarina/gocsv"
	"github.com/rs/cors"
	"github.com/rs/zerolog/log"
)

type Node struct {
	ID           int                `json:"id"`
	Humidity     float64            `json:"humidity"`
	Temperature  float64            `json:"temperature"`
	StepCount    float64            `json:"step_count"`
	StressLevel  float64            `json:"-"`
	Result       map[string]float64 `json:"-"`
	ChosenResult float64            `json:"-"`
	ClusterCode  string             `json:"-"`
}

func main() {

	cors := initCors()

	router := chi.NewRouter()
	router.Use(cors.Handler)
	router.Use(EnabledCors)

	router.Get("/ping", func(rw http.ResponseWriter, r *http.Request) {
		log.Printf("Hello from %s", r.RemoteAddr)
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte("FONG\n"))
	})

	router.Post("/kmeans", controller.KMeansController)
	router.Post("/kmedoids", controller.KMedoidsController)
	router.Post("/upload-dataset", controller.UploadDataset)
	router.Get("/datasets", controller.ListDataController)

	serve(router)
}

func initData() {
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
		//fmt.Println("node: ", mainNodes[k])
		mainNodes[k].ID = k + 1
	}
}

func serve(handler http.Handler) {
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", "", 3080),
		Handler: handler,
	}

	go onTerminate(server)

	log.Info().Msgf("Starting server at %s:%d", "", 3080)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal().Err(err).Msg("could not start server")
	}
}

func onTerminate(server *http.Server) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("terminating")
	if err := server.Shutdown(context.Background()); err != nil {
		log.Fatal().Err(err).Msg("cannot shutdown server")
	}
}

func initCors() *cors.Cors {

	cors := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"*"},
		AllowedHeaders: []string{"*"},
	})

	return cors
}

func EnabledCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		(w).Header().Set("Access-Control-Allow-Origin", "*")
		(w).Header().Set("Access-Control-Allow-Methods", "*")
		next.ServeHTTP(w, r)
	})
}
