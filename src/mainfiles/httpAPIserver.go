package mainfiles

import (
	"encoding/base64"
	"flag"
	"fmt"
	"github.com/streadway/amqp"
	"io/ioutil"
	"log"
	"net/http"
	"runtime"
	"strings"
	"time"
)

type qqconnect struct {
	qvo *rqueue
}

var myMQ = &qqconnect{}

var amqpUri = flag.String("r", to.rabbitAccess, "RabbitMQ URI")

var (
	rabbitConn       *amqp.Connection
	rabbitCloseError chan *amqp.Error
)

func connectToRabbitMQ(uri string) *amqp.Connection {
	for {
		conn, err := amqp.Dial(uri)

		if err == nil {
			return conn
		}

		log.Println(err)
		log.Printf("Trying to reconnect to RabbitMQ at %s\n", uri)
		time.Sleep(500 * time.Millisecond)
	}
}

func rabbitConnector(uri string) {
	var rabbitErr *amqp.Error

	for {
		rabbitErr = <-rabbitCloseError
		if rabbitErr != nil {
			log.Printf("Connecting to %s\n", *amqpUri)
			rabbitConn = connectToRabbitMQ(uri)
			rabbitCloseError = make(chan *amqp.Error)
			rabbitConn.NotifyClose(rabbitCloseError)
		}
	}
}

// -------------------------------------------------------------------------- //

// -------------------------------------------------------------------------- //
func dynHandler(w http.ResponseWriter, r *http.Request) {
	// -- basic auth -- //
	const unauth = http.StatusUnauthorized

	if to.serverAuth {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Basic ") {
			log.Print("Invalid authorization:", auth)
			http.Error(w, http.StatusText(unauth), unauth)
			return
		}
		up, err := base64.StdEncoding.DecodeString(auth[6:])
		if err != nil {
			log.Print("authorization decode error:", err)
			http.Error(w, http.StatusText(unauth), unauth)
			return
		}
		if string(up) != authorized["server"] {
			http.Error(w, http.StatusText(unauth), unauth)
			return
		}
	}

	// -- ---------- -- //

	switch r.Method {
	case "GET":
		myParam := r.URL.Query().Get("param")
		if myParam != "" {
			_, _ = fmt.Fprintln(w, "Param is:", myParam)

		}
		key := r.FormValue("key")
		if key != "" {
			_, _ = fmt.Fprintln(w, "Key is:", key)
		}
	case "POST":
		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}

		go func(out chan<- string) {
			out <- string(reqBody)
		}(to.queue)

	default:
		_, _ = fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
	}
}

func playmux0() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", dynHandler)

	s1 := http.Server{
		Addr:         to.httpAddress,
		Handler:      mux,
		ReadTimeout:  100 * time.Second,
		WriteTimeout: 100 * time.Second,
	}
	_ = s1.ListenAndServe()

}

func mxhandl(w http.ResponseWriter, r *http.Request) {
	mz := printStats()
	_, _ = fmt.Fprintln(w, mz)
}

func playmux1() {
	mux1 := http.NewServeMux()
	mux1.HandleFunc("/", mxhandl)

	s2 := http.Server{
		Addr:         "127.0.0.1:9191",
		Handler:      mux1,
		ReadTimeout:  100 * time.Second,
		WriteTimeout: 100 * time.Second,
	}
	_ = s2.ListenAndServe()

}

func lobi() (q *rqueue) {
	switch to.rabbitmqConnected {
	case false:
		myMQ.qvo = NewQueue(to.rabbitAccess, to.rQueueName)
		to.rabbitmqConnected = true
		return myMQ.qvo
	}
	return myMQ.qvo
}

// ---------------- //
var chocho = make(chan bool)

func RunServer() {
	setVarsik()
	http.HandleFunc("/", dynHandler)
	fmt.Println("starting server at: " + to.httpAddress)

	go playmux0()
	if to.monenabled {
		go playmux1()
	}

	log.Print("Started Proxy ")
	for j := 0; j < to.dispatchersCount; j++ {
		go func() {
			for {
				processData()
			}
		}()
	}

	p := true
	go func(in chan bool) {
		for {
			vle := <-in
			_ = vle
			if pause != p {

				switch to.internalQueue {
				case false:
					//log.Print("Started Proxy with RabbitMQ")
					switch pause {
					case true:
						log.Println("Got error from upstream server, stopping consumers")
						go runConsumer(false)
						p = pause
					case false:
						log.Println("Upstream server is up now, processing")
						go runConsumer(true)
						p = pause
					}
				}

			}
		}
	}(chocho)

	runtime.Gosched()

	//log.Println("Currently running", runtime.NumGoroutine(), "Goroutines")

	forever := make(chan bool)
	<-forever

	//log.Fatal(http.ListenAndServe(":9090", nil))

}
