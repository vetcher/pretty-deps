package internal

import (
	"fmt"
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

const (
	CLIENT   = "CLIENT"
	SERVER   = "SERVER"
	CONSUMER = "CONSUMER"
	PRODUCER = "PRODUCER"
	other    = "other"
	cs       = "client-server"
	cp       = "consumer-producer"
)

type State struct {
	Services   []string
	Links      []Link
	Begin, End time.Time
}

type Link struct {
	From   string
	To     string
	Amount uint64
	Kind   string
}

func (l *Link) fromToFillEmpty(from, to string) {
	if l.From == "" {
		l.From = from
	}
	if l.To == "" {
		l.To = to
	}
}

type tempLink struct {
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

func NewCore(endpoint string) *Core {
	return &Core{endpoint: endpoint}
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

func (c Core) GetState(begin, end time.Time) State {
	const parallelCount = 20
	c.services.mx.Lock()
	defer c.services.mx.Unlock()
	var wg sync.WaitGroup
	resDelivery := make(chan tempLink, parallelCount)
	state := make(chan State)
	go func() {
		cc := make([]string, len(c.services.list))
		copy(cc, c.services.list)
		result := State{Services: cc, Begin: begin, End: end}
		for chunk := range resDelivery {
			result.Links = mergeLinks(result.Links, chunk.links)
		}
		state <- result
	}()
	semaphore := make(chan struct{}, parallelCount)
	for _, name := range c.services.list {
		semaphore <- struct{}{}
		wg.Add(1)
		go func(name string) {
			ll, err := c.getLinks(
				fmt.Sprintf("%s%s?serviceName=%s&limit=%d", c.endpoint, tracesPath, name, 100),
			)
			if err != nil {
				fmt.Println(name, err)
			} else {
				resDelivery <- ll
			}
			<-semaphore
			wg.Done()
		}(name)
	}
	wg.Wait()
	close(semaphore)
	close(resDelivery)
	return <-state
}

func mergeLinks(src []Link, newChunk []Link) []Link {
ExtLoop:
	for i := range newChunk {
		kind := getNormalizedKind(newChunk[i].Kind)
		for j := range src {
			if src[j].From == newChunk[i].From && src[j].To == newChunk[i].To && src[j].Kind == kind {
				src[j].Amount += newChunk[i].Amount
				continue ExtLoop
			}
		}
		newChunk[i].Kind = kind
		src = append(src, newChunk[i])
	}
	return src
}

func getNormalizedKind(a string) string {
	switch a {
	case CLIENT, SERVER:
		return cs
	case CONSUMER, PRODUCER:
		return cp
	default:
		return other
	}
}

type pair struct{ From, To string }

func (p *pair) fillNotEmpty(from, to string) {
	if from != "" {
		p.From = from
	}
	if to != "" {
		p.To = to
	}
}

func (c Core) getLinks(targetUrl string) (tl tempLink, err error) {
	resp, err := http.Get(targetUrl)
	if err != nil {
		return
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	p := c.ppool.Get()
	defer c.ppool.Put(p)
	root, err := p.ParseBytes(data)
	if err != nil {
		return tl, fmt.Errorf("root: %v", err)
	}
	tl.links = make([]Link, 0, 20)
	pairs := make(map[string]pair)
	for _, trace := range root.GetArray() {
		for _, span := range trace.GetArray() {
			kind := string(span.GetStringBytes("kind"))
			var from, to string
			switch kind {
			case "CLIENT":
				from = string(span.GetStringBytes("localEndpoint", "serviceName"))
				to = string(span.GetStringBytes("remoteEndpoint", "serviceName"))
			case "SERVER":
				to = string(span.GetStringBytes("localEndpoint", "serviceName"))
				from = string(span.GetStringBytes("remoteEndpoint", "serviceName"))
			case "PRODUCER":
				from = string(span.GetStringBytes("localEndpoint", "serviceName"))
				to = MessageBroker
			case "CONSUMER":
				to = string(span.GetStringBytes("localEndpoint", "serviceName"))
				from = MessageBroker
			}
			if from == "" || to == "" {
				// Try to find pair if option ClientServerSameSpan is on
				id := string(span.GetStringBytes("id"))
				p := pairs[id]
				p.fillNotEmpty(from, to)
				pairs[id] = p
				if p.From == "" || p.To == "" {
					continue
				}
				from, to = p.From, p.To
			}
			link, ll := findLink(from, to, kind, tl.links)
			link.Amount++
			tl.links = ll
		}
	}
	return
}

func findLink(from, to, kind string, ll []Link) (*Link, []Link) {
	for i := range ll {
		if ll[i].From == from && ll[i].To == to && ll[i].Kind == kind {
			return &ll[i], ll
		}
	}
	ll = append(ll, Link{Kind: kind, From: from, To: to})
	return &ll[len(ll)-1], ll
}
