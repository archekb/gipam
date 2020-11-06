package config

import (
	"errors"
	"flag"
	"os"
	"strconv"
)

// New - create new config from defaults, env and flags, after than check parameter on exists. No validate, only exists!
func New() (*Config, error) {
	var cnf Config = Config{}

	cnf.setDefaults()
	cnf.parseEnv()
	cnf.parceFlags()

	err := cnf.Check()
	if err != nil {
		return nil, err
	}

	return &cnf, nil
}

// Config - contains Server and Leases config
type Config struct {
	Server struct {
		Address string
	}

	Lease struct {
		File   string
		IPv6   string
		IPv6AB uint
		IPv4   string
		IPv4AB uint
	}
}

func (cnf *Config) setDefaults() {
	cnf.Lease.File = "lease.json"
	cnf.Lease.IPv6 = ""
	cnf.Lease.IPv6AB = 64
	cnf.Lease.IPv4 = ""
	cnf.Lease.IPv4AB = 24
}

func (cnf *Config) parseEnv() {
	// Server config
	cnf.Server.Address = getEnvParam("GIPAM_ADDRESS", cnf.Server.Address).(string)

	// Lease config
	cnf.Lease.File = getEnvParam("GIPAM_FILE", cnf.Lease.File).(string)
	cnf.Lease.IPv6 = getEnvParam("GIPAM_V6", cnf.Lease.IPv6).(string)
	cnf.Lease.IPv6AB = getEnvParam("GIPAM_V6AB", cnf.Lease.IPv6AB).(uint)
	cnf.Lease.IPv4 = getEnvParam("GIPAM_V4", cnf.Lease.IPv4).(string)
	cnf.Lease.IPv4AB = getEnvParam("GIPAM_V4AB", cnf.Lease.IPv4AB).(uint)
}

func (cnf *Config) parceFlags() {
	// Server config
	flag.StringVar(&cnf.Server.Address, "address", cnf.Server.Address, "Server address and port to listen 'host:port' or ':port'. If empty used UNIX socket.")

	// Lease config
	flag.StringVar(&cnf.Lease.File, "file", cnf.Lease.File, "Lease file uses for save state to file and restore it after restart")
	flag.StringVar(&cnf.Lease.IPv6, "v6", cnf.Lease.IPv6, "Main IPv6 address pool. Example: fe80::/56")
	flag.UintVar(&cnf.Lease.IPv6AB, "v6ab", cnf.Lease.IPv6AB, "Mask of IPv6 allocated block. Example: 64")
	flag.StringVar(&cnf.Lease.IPv4, "v4", cnf.Lease.IPv4, "Main IPv4 address pool. Example: 192.168.0.0/16")
	flag.UintVar(&cnf.Lease.IPv4AB, "v4ab", cnf.Lease.IPv4AB, "Mask of IPv4 allocated block. Example: 24")

	flag.Parse()
}

// Check - check config parametrs
func (cnf *Config) Check() error {
	// no ip alocated blocks and no leases filename (or file must be Wipe)
	if cnf.Lease.IPv6 == "" && cnf.Lease.IPv4 == "" && cnf.Lease.File == "" {
		return errors.New("No leases configuration")
	}

	return nil
}

// Help - print flag defaults
func Help() {
	flag.PrintDefaults()
}

// utils
func getEnvParam(name string, def interface{}) interface{} {
	env := os.Getenv(name)
	if env == "" {
		return def
	}

	switch def.(type) {
	case string:
		return env

	case int:
		if r, err := strconv.Atoi(env); err == nil {
			return r
		}
		return def

	case uint:
		if r, err := strconv.ParseUint(env, 10, 0); err == nil { // return uint with default of system bitsize 32 or 64
			return r
		}
		return def

	case bool:
		if r, err := strconv.ParseBool(env); err == nil {
			return r
		}
		return def
	}

	return def
}
