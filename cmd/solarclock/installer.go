package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"
)

func exePath() string {
	e := os.Args[0]
	p, err := filepath.Abs(e)
	if err != nil {
		panic(err)
	}
	return p
}

func configureService(cfg func(*mgr.Mgr) error) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	return cfg(m)
}

func createService() error {
	return configureService(createServiceOption)
}

func removeService() error {
	return configureService(removeServiceOption)
}

func stopService() error {
	return configureService(stopServiceOption)
}

func startService() error {
	return configureService(startServiceOption)
}

func createServiceOption(m *mgr.Mgr) error {
	fmt.Fprintln(os.Stdout, "creating service", serviceName)
	p := exePath()
	if s, err := m.OpenService(serviceName); err == nil {
		fmt.Fprintln(os.Stdout, "service exists", serviceName)
		s.Close()
		return nil
	}
	s, err := m.CreateService(serviceName, p, mgr.Config{
		StartType:   mgr.StartAutomatic,
		Description: serviceDesc,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "error creating service", err)
		return err
	}
	defer s.Close()
	err = eventlog.InstallAsEventCreate(serviceName, eventlog.Error|eventlog.Warning|eventlog.Info)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error installing event source", err)
		s.Delete()
		return err
	}
	return nil
}

func removeServiceOption(m *mgr.Mgr) error {
	fmt.Fprintln(os.Stdout, "removing service", serviceName)
	s, err := m.OpenService(serviceName)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error opening service", err)
		return nil
	}
	defer s.Close()
	sts, err := s.Query()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error querying service", err)
		return err
	}
	if sts.State == svc.Running {
		fmt.Fprintln(os.Stdout, " stopping service")
		_, err := s.Control(svc.Stop)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error stopping service", err)
			return err
		}
	}
	err = s.Delete()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error deleting service", err)
		return err
	}
	err = eventlog.Remove(serviceName)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error removing event log", err)
		return err
	}
	return nil
}

func stopServiceOption(m *mgr.Mgr) error {
	fmt.Fprintln(os.Stdout, "stopping service", serviceName)
	s, err := m.OpenService(serviceName)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error getting service", err)
		return nil
	}
	defer s.Close()
	sts, err := s.Query()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error getting service status", err)
		return err
	}
	if sts.State == svc.Stopped {
		fmt.Fprintln(os.Stdout, "Service already stopped")
		return nil
	}
	sts, err = s.Control(svc.Stop)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error stopping service", err)
		return err
	}
	if sts.State != svc.Stopped {
		fmt.Fprintln(os.Stdout, "Status", getStateText(sts.State))
		time.Sleep(100 * time.Millisecond)
		sts, _ = s.Query()
	}
	fmt.Fprintln(os.Stdout, "Status", getStateText(sts.State))
	return nil
}

func startServiceOption(m *mgr.Mgr) error {
	fmt.Fprintln(os.Stdout, "starting service", serviceName)
	s, err := m.OpenService(serviceName)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error getting service", err)
		return nil
	}
	defer s.Close()
	sts, err := s.Query()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error getting service status", err)
		return err
	}
	if sts.State == svc.Running {
		fmt.Fprintln(os.Stdout, "Status", getStateText(sts.State))
	}
	err = s.Start()
	if err != nil {
		return err
	}
	sts, err = s.Query()
	if err != nil {
		return err
	}
	if sts.State == svc.StartPending {
		fmt.Fprintln(os.Stdout, "Status", getStateText(sts.State))
		time.Sleep(100 * time.Millisecond)
		sts, _ = s.Query()
	}
	fmt.Fprintln(os.Stdout, "Status", getStateText(sts.State))
	return nil
}

func getStateText(s svc.State) string {
	switch s {
	case svc.Stopped:
		return "Stopped"
	case svc.StartPending:
		return "StartPending"
	case svc.StopPending:
		return "StopPending"
	case svc.Running:
		return "Running"
	case svc.ContinuePending:
		return "ContinuePending"
	case svc.PausePending:
		return "PausePending"
	case svc.Paused:
		return "Paused"
	}
	return "Other"
}
