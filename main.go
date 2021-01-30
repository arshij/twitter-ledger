package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/davecgh/go-spew/spew" // allows us to view structs and slices cleanly formatted in our console
	"github.com/dghubble/oauth1"      // allows us to make client calls to Twitter
	"github.com/gorilla/mux"          // used for writing web handlers
	"github.com/joho/godotenv"        // used to manage environment variables
)

// defines and creates routes
func makeMuxRouter() http.Handler {
	muxRouter := mux.NewRouter()

	muxRouter.HandleFunc("/user/{id}", handleGetTweets).Methods("GET")
	muxRouter.HandleFunc("/blockchain", handleGetBlockchain).Methods("GET")
	muxRouter.HandleFunc("/blockchain/{id}", handleWriteBlock).Methods("POST")
	return muxRouter
}

func respondWithError(err error, w http.ResponseWriter) {
	log.Println(err)
	w.WriteHeader(500)
	w.Write([]byte(err.Error()))
}

// web server set up
func run() error {
	// Load Twitter API credentials
	config = oauth1.NewConfig(os.Getenv("APIKEY"), os.Getenv("APISECRET"))
	token = oauth1.NewToken(os.Getenv("TOKEN"), os.Getenv("TOKENSECRET"))

	mux := makeMuxRouter()
	httpAddr := os.Getenv("PORT")
	log.Println("Listening on ", os.Getenv("PORT"))
	s := &http.Server{
		Addr:           ":" + httpAddr,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	if err := s.ListenAndServe(); err != nil {
		return err
	}

	return nil
}

func main() {
	err := godotenv.Load() // allows us to read in variables from .env file
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		genesisBlock := Block{0, "", "", ""} // initial/first Block
		spew.Dump(genesisBlock)
		Blockchain = append(Blockchain, genesisBlock)
	}()
	log.Fatal(run())

}
