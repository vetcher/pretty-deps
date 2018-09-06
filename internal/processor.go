package internal

import (
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/valyala/fastjson"
)

const (
	servicesPath = "/services"
	tracesPath   = "/traces"
)

type State struct {
	Services []string
	Links    map[string][]Link
}

type Link struct {
	To     string
	Amount uint64
	Type   string
}

type tempLink struct {
	name  string
	links []Link
}

type Core struct {
	endpoint string
	ppool    fastjson.ParserPool

	services struct {
		lastCheckOut time.Time
		list         []string
		mx           sync.Mutex
	}
}

func (c *Core) UpdateServicesList() error {
	c.services.mx.Lock()
	defer c.services.mx.Unlock()
	resp, err := http.Get(c.endpoint + servicesPath)
	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	p := c.ppool.Get()
	defer c.ppool.Put(p)
	list, err := p.ParseBytes(body)
	if err != nil {
		return err
	}
	arr := list.GetArray()
	result := make([]string, len(arr))
	for i := range arr {
		result[i] = string(arr[i].GetStringBytes())
	}
	c.services.list = result
	c.services.lastCheckOut = time.Now()
	return nil
}

func (c Core) GetState(from, to time.Time) State {
	const parallelCount = 20
	c.services.mx.Lock()
	defer c.services.mx.Unlock()
	var wg sync.WaitGroup
	resDelivery := make(chan tempLink, parallelCount)
	state := make(chan State)
	go func() {
		cc := make([]string, len(c.services.list))
		copy(cc, c.services.list)
		result := State{Services: cc, Links: make(map[string][]Link, len(cc))}
		for chunk := range resDelivery {
			result.Links[chunk.name] = chunk.links
		}
		state <- result
	}()
	semaphore := make(chan struct{}, parallelCount)
	for _, name := range c.services.list {
		semaphore <- struct{}{}
		wg.Add(1)
		go func(name string) {
			resDelivery <- c.getLinks(name)
			<-semaphore
			wg.Done()
		}(name)
	}
	wg.Wait()
	close(semaphore)
	close(resDelivery)
	return <-state
}

func (c Core) getLinks(name string) (tl tempLink, err error) {
	resp, err := http.Get(c.endpoint + tracesPath + "?serviceName=" + name)
	if err != nil {
		return
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	tl.name = name
	p := c.ppool.Get()
	defer c.ppool.Put(p)
	root, err := p.ParseBytes(data)
	if err != nil {
		return
	}
	tl.links = make([]Link, 0, 4) // [ CLIENT, SERVER, PRODUCER, CONSUMER ]
	for _, trace := range root.GetArray() {
		for _, span := range trace.GetArray() {
			kind := string(span.GetStringBytes("kind"))
			link := findByKind(kind, tl.links)
			link.Amount++
		}
	}
	return
}

func findByKind(kind string, ll []Link) *Link {
	for i := range ll {
		if ll[i].Type == kind {
			return &ll[i]
		}
	}
	ll = append(ll, Link{Type: kind})
	return &ll[len(ll)-1]
}
