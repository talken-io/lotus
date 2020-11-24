package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"

	"github.com/filecoin-project/lotus/chain/types"
	"github.com/gorilla/mux"

	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("drand")

type Transaction struct {
	Message *types.Message `json:"message"`
}

func convertMessage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var transaction Transaction
	_ = json.NewDecoder(r.Body).Decode(&transaction)
	message := transaction.Message

	mb, err := message.ToStorageBlock()
	if err != nil {
		json.NewEncoder(w).Encode(err)
	} else {
		json.NewEncoder(w).Encode(mb.RawData())
	}
}

func main() {
	ip := flag.String("ip", "127.0.0.1", "Listening IP")
	port := flag.Int("port", 3030, "Listening Port")
	flag.Parse()

	r := mux.NewRouter()
	r.HandleFunc("/convertMessage", convertMessage).Methods("POST")

	server := fmt.Sprintf("%s:%d", *ip, *port)
	log.Fatal(http.ListenAndServe(server, r))
}
