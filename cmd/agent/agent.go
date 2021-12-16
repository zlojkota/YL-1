package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/smtp"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/labstack/gommon/log"
	"github.com/zlojkota/YL-1/internal/agentcollector"
	"github.com/zlojkota/YL-1/internal/collector"
)

type Worker struct {
	ServerAddr     *string        `env:"ADDRESS" envDefault:"127.0.0.1:8080"`
	ReportInterval *time.Duration `env:"REPORT_INTERVAL" envDefault:"10s"`
	PoolInterval   *time.Duration `env:"POLL_INTERVAL" envDefault:"2s"`
	HashKey        *string        `env:"KEY" envDefault:""`
	sendmsg        string
}

func (p *Worker) RequestServe(req *http.Request) {

	body, _ := io.ReadAll(req.Body)
	p.sendmsg = p.sendmsg +
		"LOG: url  -" + req.URL.String() + "\r\n" +
		"LOG: body -" + string(body) + "\r\n--------------------------\r\n"

	fmt.Println("Email Sent Successfully!")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Error(err)
		return
	}
	defer res.Body.Close()
	log.Print(req.URL.Path)
}

func main() {

	var t collector.Collector
	var agent agentcollector.Agent
	var worker Worker

	err := env.Parse(&worker)
	if err != nil {
		log.Fatal(err)
	}

	if _, ok := os.LookupEnv("ADDRESS"); !ok {
		worker.ServerAddr = flag.String("a", "127.0.0.1:8080", "ADDRESS")
	} else {
		_ = flag.String("a", "127.0.0.1:8080", "ADDRESS")
	}
	if _, ok := os.LookupEnv("REPORT_INTERVAL"); !ok {
		worker.ReportInterval = flag.Duration("r", 10*time.Second, "REPORT_INTERVAL")
	} else {
		_ = flag.Duration("r", 10*time.Second, "REPORT_INTERVAL")
	}
	if _, ok := os.LookupEnv("POLL_INTERVAL"); !ok {
		worker.PoolInterval = flag.Duration("p", 2*time.Second, "POLL_INTERVAL")
	} else {
		_ = flag.Duration("p", 2*time.Second, "POLL_INTERVAL")
	}
	if _, ok := os.LookupEnv("KEY"); !ok {
		worker.HashKey = flag.String("k", "", "KEY")
	} else {
		_ = flag.String("k", "", "POLL_INTERVAL")
	}
	flag.Parse()

	a, _ := json.Marshal(worker)
	log.Error(string(a))

	t.Handle(*worker.PoolInterval, &agent)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		<-sigChan
		log.Error("Stopping")
		from := "ololoevqwerty799@gmail.com"
		password := "zlojKOTA123"
		to := []string{
			"ololoevqwerty799@gmail.com",
		}
		smtpHost := "smtp.gmail.com"
		smtpPort := "587"
		auth := smtp.PlainAuth("", from, password, smtpHost)
		msg := []byte("To: recipient@example.net\r\n" +
			"Subject: Logs for " + string(a) + "\r\n" +
			"\r\n" +
			"This is the email body.\r\n")
		err = smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, msg)
		if err != nil {
			fmt.Println(err)
			return
		}
		t.Done <- true
	}()
	t.Run()

}
