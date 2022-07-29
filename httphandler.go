package httphandler

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

const maxConnection = 999

type HTTPHandler struct {
	mx                 *sync.Mutex
	countOfConnections int
}

func NewHTTPHandler() *HTTPHandler {
	return &HTTPHandler{
		mx:                 &sync.Mutex{},
		countOfConnections: 0,
	}
}

func (h *HTTPHandler) IncrementConnections() {
	h.mx.Lock()
	defer h.mx.Unlock()
	h.countOfConnections++
}

func (h *HTTPHandler) DecrementConnections() {
	h.mx.Lock()
	defer h.mx.Unlock()
	h.countOfConnections--
}

func (h *HTTPHandler) GetConnectionCount() int {
	h.mx.Lock()
	defer h.mx.Unlock()
	return h.countOfConnections
}

func (h *HTTPHandler) IncrementAndGetConnections() int {
	h.mx.Lock()
	defer h.mx.Unlock()
	h.countOfConnections++
	return h.countOfConnections
}

func (h *HTTPHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if h.IncrementAndGetConnections() > maxConnection {
		rw.WriteHeader(http.StatusServiceUnavailable)
		_, err := rw.Write([]byte("too many connections"))
		if err != nil {
			log.Fatalln(err.Error())
		}
		h.DecrementConnections()
		return
	}
	defer h.DecrementConnections()

	if req.Method != http.MethodPost {
		rw.WriteHeader(http.StatusNotFound)
		_, err := rw.Write([]byte(fmt.Sprintf("method %s is not allowed", req.Method)))
		if err != nil {
			log.Fatalln(err.Error())
		}
		return
	}

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		_, err := rw.Write([]byte(err.Error()))
		if err != nil {
			log.Fatalln(err.Error())
		}
		return
	}
	urls := strings.Split(string(body), "\n")
	var resultStings []string

	var wg sync.WaitGroup
	for u := range urls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			contentLength, err := sendRequestAndGetContentLength(url)
			if err != nil {
				log.Println(err.Error())
			}
			resultStings = append(resultStings, strconv.Itoa(contentLength))
		}(urls[u])
	}
	wg.Wait()

	_, err = rw.Write([]byte(strings.Join(resultStings, "\n")))
	if err != nil {
		log.Fatal(err.Error())
		return
	}
}

func sendRequestAndGetContentLength(url string) (int, error) {
	response, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer func() {
		err := response.Body.Close()
		if err != nil {
			log.Fatal(err.Error())
		}
	}()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return 0, err
	}
	return len(body), nil
}
