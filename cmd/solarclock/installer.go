package main

import (
	"fmt"
	"os"
	"path/filepath"

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
	p := exePath()
	if s, err := m.OpenService(serviceName); err == nil {
		s.Close()
		return nil
	}
	s, err := m.CreateService(serviceName, p, mgr.Config{
		StartType:   mgr.StartAutomatic,
		Description: serviceDesc,
	})
	if err != nil {
		return err
	}
	defer s.Close()
	err = eventlog.InstallAsEventCreate(serviceName, eventlog.Error|eventlog.Warning|eventlog.Info)
	if err != nil {
		s.Delete()
		return err
	}
	return nil
}

func removeServiceOption(m *mgr.Mgr) error {
	fmt.Fprintln(os.Stderr, "removing service", serviceName)
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
	s, err := m.OpenService(serviceName)
	if err != nil {
		return nil
	}
	defer s.Close()
	sts, err := s.Query()
	if err != nil {
		return err
	}
	if sts.State == svc.Running {
		s.Control(svc.Stop)
	}
	return nil
}

func startServiceOption(m *mgr.Mgr) error {
	s, err := m.OpenService(serviceName)
	if err != nil {
		return nil
	}
	defer s.Close()
	sts, err := s.Query()
	if err != nil {
		return err
	}
	if sts.State != svc.Running {
		s.Start()
	}
	return nil
}
