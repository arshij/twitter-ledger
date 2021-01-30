package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"

	"github.com/davecgh/go-spew/spew" // allows us to view structs and slices cleanly formatted in our console
  "github.com/gorilla/mux"          // used for writing web handlers
)

// Block represents each 'item' in the blockchain
type Block struct {
	Index 			int 		// position of the data record in the blockchain
	TweetHash   string	// hash representation of user's original tweets
	Hash      	string 	// SHA256 identifier representing this record
	PrevHash  	string 	// SHA256 identifier of the previous block
}

var Blockchain []Block

// concatenates elements of a Block and returns the SHA256 hash as a string
func calculateHash(block Block) string {
	record := string(block.Index) + string(block.TweetHash) + block.PrevHash
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

// builds and returns a new Block object
func generateBlock(oldBlock Block, TweetHash string) (Block, error) {
	var newBlock Block

	newBlock.Index = oldBlock.Index + 1
	newBlock.TweetHash = TweetHash
	newBlock.PrevHash = oldBlock.Hash
	newBlock.Hash = calculateHash(newBlock)

	return newBlock, nil
}

// ensures the block has not been tampered with
func isBlockValid(newBlock, oldBlock Block) bool {
	// ensure Index is incremented as expected
	if oldBlock.Index+1 != newBlock.Index {
		return false
	}

	// ensure the hashes match up
	if oldBlock.Hash != newBlock.PrevHash {
		return false
	}

	// double check the hash of the current Block
	if calculateHash(newBlock) != newBlock.Hash {
		return false
	}

	return true
}

// in the event that 2 nodes add 2 different Blocks at the same time, validate the
// Blockchain that has more Blocks
func replaceChain(newBlocks []Block) {
	if len(newBlocks) > len(Blockchain) {
		Blockchain = newBlocks
	}
}

// GET handler to retrieve the Blockchain
func handleGetBlockchain(w http.ResponseWriter, r *http.Request) {
	bytes, err := json.MarshalIndent(Blockchain, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	io.WriteString(w, string(bytes))
}

// POST handler to write to the Blockchain
func handleWriteBlock(w http.ResponseWriter, r *http.Request) {
  vars := mux.Vars(r)
	userID := vars["id"]	// Twitter user handle
  tweetHash := calculateTweetHash(userID)

	defer r.Body.Close()

	// create new Block
	newBlock, err := generateBlock(Blockchain[len(Blockchain)-1], tweetHash)
	if err != nil {
		respondWithJSON(w, r, http.StatusInternalServerError, tweetHash)
		return
	}

	// quick sanity check
	if isBlockValid(newBlock, Blockchain[len(Blockchain)-1]) {
		newBlockchain := append(Blockchain, newBlock)
		replaceChain(newBlockchain) // here to simulate a new peer entering the network that will need to download the whole Blockchain
		spew.Dump(Blockchain)       // convenient function that pretty prints our structs into the console, useful for debugging.
	}

	respondWithJSON(w, r, http.StatusCreated, newBlock)
}

// wrapper function to alert us of status
func respondWithJSON(w http.ResponseWriter, r *http.Request, code int, payload interface{}) {
	response, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("HTTP 500: Internal Server Error"))
		return
	}
	w.WriteHeader(code)
	w.Write(response)
}
