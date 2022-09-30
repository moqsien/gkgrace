package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/moqsien/gkgrace"
	"github.com/moqsien/gkgrace/apps/xgin"
	"github.com/moqsien/processes/signals"
)

func run() {
	grace := gkgrace.New()
	gin.SetMode(gin.ReleaseMode)
	app := xgin.New()
	app.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "OK! Hello app!")
	})
	app.SetAddr(&gkgrace.Address{
		Network: "tcp",
		Host:    "0.0.0.0",
		Port:    8080,
	})
	grace.Register(app)
	go app.Run()

	app1 := xgin.New()
	app1.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "OK! Hello app1!")
	})
	app1.SetAddr(&gkgrace.Address{
		Network: "tcp",
		Host:    "0.0.0.0",
		Port:    8081,
	})
	grace.Register(app1)
	go app1.Run()

	grace.Wait()
}

func main() {
	fmt.Println("Current pid: ", os.Getpid())
	if len(os.Args) < 2 {
		run()
	} else {
		pid, _ := strconv.Atoi(os.Args[1])
		if pid != 0 {
			// send restart signal to process
			_ = signals.KillPid(pid, signals.ToSignal("SIGUSR2"), false)
		}
	}
}
