package main

import (
	"flag"
	"net/http"
	"os"
	"os/exec"
	"path"
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
		groups := ctx.QueryArray("group")
		err = core.UpdateServicesList()
		if err != nil {
			return
		}
		now := time.Now()
		state := core.GetState(now.AddDate(0, 0, -1), now)
		dotData, err := internal.StateToGraph(state, groups...)
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
