//go:build windows

package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/pachecot/solarclock/solartime"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
)

const serviceName = "SolarClockService"
const serviceDesc = "Solar Clock web Service"
const serviceParametersKey = `SYSTEM\CurrentControlSet\Services\` + serviceName + `\Parameters`

type solarClockService struct {
	elog   debug.Log
	port   uint32
	cancel context.CancelFunc
	ctx    context.Context
	srv    *http.Server
}

func createClockService(port uint32, elog debug.Log) *solarClockService {
	ctx, cancel := context.WithCancel(context.Background())
	srv := &http.Server{Addr: fmt.Sprintf(":%d", port)}
	return &solarClockService{
		port:   port,
		ctx:    ctx,
		cancel: cancel,
		srv:    srv,
		elog:   elog,
	}
}

func (s *solarClockService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {

	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown

	s.elog.Info(1, fmt.Sprintf("%s service state starting", serviceName))
	changes <- svc.Status{State: svc.StartPending}

	go func() {
		s.startHttpServer()
	}()

	s.elog.Info(1, fmt.Sprintf("%s service status running", serviceName))
	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

loop:
	for {
		select {
		case <-s.ctx.Done():
			s.elog.Info(1, fmt.Sprintf("%s ctx done", serviceName))
			break loop

		case c := <-r:
			switch c.Cmd {

			case svc.Interrogate:
				changes <- c.CurrentStatus
				time.Sleep(100 * time.Millisecond)
				changes <- c.CurrentStatus

			case svc.Stop, svc.Shutdown:
				testOutput := strings.Join(args, "-")
				testOutput += fmt.Sprintf("-%d", c.Context)
				s.elog.Info(1, testOutput)
				s.srv.Shutdown(context.Background())
				break loop

			case svc.Continue:
				changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

			default:
				s.elog.Error(1, fmt.Sprintf("unexpected control request #%d", c))
			}
		}
	}
	changes <- svc.Status{State: svc.StopPending}
	return
}

func (s *solarClockService) startHttpServer() {

	http.HandleFunc("/xml", solartime.XmlHandler)
	http.HandleFunc("/json", solartime.JsonHandler)

	s.elog.Info(1, fmt.Sprintf("%s http server starting on %s", serviceName, s.srv.Addr))
	err := s.srv.ListenAndServe()
	if err != http.ErrServerClosed {
		s.elog.Info(1, fmt.Sprintf("%s http server stopped %v", serviceName, err))
	}
	if err != nil {
		s.elog.Info(1, fmt.Sprintf("%s http server exited %v", serviceName, err))
	}
	s.cancel()
}

func (s *solarClockService) run() {
	s.elog.Info(1, fmt.Sprintf("starting %s service", serviceName))

	err := svc.Run(serviceName, s)
	if err != nil {
		s.elog.Error(1, fmt.Sprintf("%s service failed: %v", serviceName, err))
		return
	}
	s.elog.Info(1, fmt.Sprintf("%s service stopped", serviceName))
}

func (s *solarClockService) runLocal() {
	s.elog.Info(1, fmt.Sprintf("starting %s service", serviceName))

	s.startHttpServer()
	fmt.Println("service stopped")
}
