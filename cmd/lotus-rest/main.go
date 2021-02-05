package main

import (
	"encoding/json"
	"flag"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/filecoin-project/lotus/chain/types"
	"github.com/gorilla/mux"

	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("rest")

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

type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}

func ListenAndServe(addr string, num int, r http.Handler) error {
	if addr == "" {
		addr = ":http"
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	for i := 0; i < num; i++ {
		wg.Add(1)
		go func(i int) {
			log.Warnf("listener number", i)
			log.Fatal(http.Serve(tcpKeepAliveListener{ln.(*net.TCPListener)}, r))

			wg.Done()
		}(i)
	}
	wg.Wait()
	return nil
}

func main() {
	ip := flag.String("ip", "127.0.0.1", "Listening IP")
	port := flag.Int("port", 3030, "Listening Port")
	flag.Parse()

	r := mux.NewRouter()
	r.HandleFunc("/convertMessage", convertMessage).Methods("POST")

	//server := fmt.Sprintf("%s:%d", *ip, *port)
	//log.Fatal(http.ListenAndServe(server, r))
	log.Fatal(http.ListenAndServe(*ip+":"+strconv.Itoa(*port), r))

	//// Multi-Thread
	//num := runtime.NumCPU()
	//runtime.GOMAXPROCS(num)
	//log.Fatal(ListenAndServe(*ip + ":" + strconv.Itoa(*port), num, r))
}
