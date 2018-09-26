package main

import (
	"flag"
	"math"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/intel-go/bytebuf"
	"github.com/vetcher/pretty-deps/internal"
)

var (
	bindAddr   = flag.String("addr", "localhost:9000", "")
	zipkinAddr = flag.String("zipkin", "http://localhost:9411/api/v2", "")
	staticPath = flag.String("static", "./static/", "")
)

func init() {
	flag.Parse()
}

func main() {
	core := internal.NewCore(*zipkinAddr)
	r := gin.Default()
	r.Use(cors.Default())
	r.StaticFile("/", path.Join(*staticPath, "index.html"))
	r.GET("/api", func(ctx *gin.Context) {
		var err error
		defer func() {
			if err != nil {
				ctx.Status(http.StatusInternalServerError)
				ctx.Writer.WriteString(err.Error())
			}
		}()
		err = core.UpdateServicesList()
		if err != nil {
			return
		}
		now := time.Now()
		state := core.GetState(now.AddDate(0, 0, -1), now)
		params := fillVisualizationParams(ctx)
		dotData, err := internal.StateToGraph(state, params)
		if err != nil {
			return
		}
		if ctx.Query("code") == "true" {
			ctx.Writer.Write(dotData)
			return
		}
		engine, format := sthOrDef(ctx.Query("engine"), "dot"), sthOrDef(ctx.Query("format"), "svg")
		dotData, err = execGraphviz(engine, format, dotData)
		if err != nil {
			return
		}
		ctx.Writer.Write(dotData)
	})
	r.Static("/static", *staticPath)
	r.Run(*bindAddr)
}

func fillVisualizationParams(ctx *gin.Context) internal.VisualizationParams {
	params := internal.VisualizationParams{StylingNodes: make(map[string]internal.StylingParams)}
	params.Prefixes = ctx.QueryArray("group")
	params.RemovePrefix, _ = strconv.ParseBool(ctx.Query("remove-prefix"))
	params.DetailThreshold, _ = strconv.Atoi(ctx.Query("detail-threshold"))
	if params.DetailThreshold == 0 {
		params.DetailThreshold = math.MaxInt32
	}
	nodeStylesMap := ctx.QueryMap("node-styles")
	for k, v := range nodeStylesMap {
		params.StylingNodes[k] = makeMapFromString(v)
	}
	return params
}

func makeMapFromString(s string) map[string]string {
	kvs := strings.Split(s, ",")
	m := make(map[string]string, len(kvs))
	for i := range kvs {
		kv := strings.Split(kvs[i], "=")
		if len(kv) == 0 {
			continue
		}
		m[kv[0]] = strings.Join(kv[1:], "=")
	}
	return m
}

func execGraphviz(name, format string, b []byte) ([]byte, error) {
	res, src := bytebuf.New(), bytebuf.NewBuffer(b)
	cmd := exec.Command(name, "-T", format)
	cmd.Stderr = os.Stderr
	cmd.Stdout, cmd.Stdin = &res, src
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	return res.Bytes(), nil
}

func sthOrDef(s string, def string) string {
	if s == "" {
		return def
	}
	return s
}
