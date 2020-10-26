package config

import (
	"errors"
	"flag"
	"os"
	"strconv"
)

// NewConfig - create new config from defaults, env and flags, than check parameter exists. No validate, only exists!
func NewConfig() (*Config, error) {
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
		UnixSocket bool
		Address    string
	}

	Lease struct {
		File   string
		Wipe   bool
		IPv6   string
		IPv6AB uint
		IPv4   string
		IPv4AB uint
	}
}

func (cnf *Config) setDefaults() {
	// Server config
	cnf.Server.UnixSocket = false
	cnf.Server.Address = ":9090"

	// Lease config
	cnf.Lease.File = "lease.json"
	cnf.Lease.Wipe = false
	cnf.Lease.IPv6 = ""
	cnf.Lease.IPv6AB = 64
	cnf.Lease.IPv4 = ""
	cnf.Lease.IPv4AB = 24
}

func (cnf *Config) parseEnv() {
	// Server config
	cnf.Server.UnixSocket = getEnvParam("GIPAM_UNIX", cnf.Server.UnixSocket).(bool)
	cnf.Server.Address = getEnvParam("GIPAM_ADDRESS", cnf.Server.Address).(string)

	// Lease config
	cnf.Lease.File = getEnvParam("GIPAM_LF", cnf.Lease.File).(string)
	cnf.Lease.Wipe = getEnvParam("GIPAM_LFWIPE", cnf.Lease.Wipe).(bool)
	cnf.Lease.IPv6 = getEnvParam("GIPAM_V6", cnf.Lease.IPv6).(string)
	cnf.Lease.IPv6AB = getEnvParam("GIPAM_V6AB", cnf.Lease.IPv6AB).(uint)
	cnf.Lease.IPv4 = getEnvParam("GIPAM_V4", cnf.Lease.IPv4).(string)
	cnf.Lease.IPv4AB = getEnvParam("GIPAM_V4AB", cnf.Lease.IPv4AB).(uint)
}

func (cnf *Config) parceFlags() {
	// Server config
	flag.BoolVar(&cnf.Server.UnixSocket, "unix", cnf.Server.UnixSocket, "Use UNIX socket if true, else tcp connection (default false)")
	flag.StringVar(&cnf.Server.Address, "address", cnf.Server.Address, "Server address and port to listen 'host:port' or ':port'")

	// Lease config
	flag.StringVar(&cnf.Lease.File, "lf", cnf.Lease.File, "Lease file uses for save state to file and restore it after restart")
	flag.BoolVar(&cnf.Lease.Wipe, "lfwipe", cnf.Lease.Wipe, "Lease file must be wipe before starting (default false)")
	flag.StringVar(&cnf.Lease.IPv6, "v6", cnf.Lease.IPv6, "Main IPv6 address pool. Example: fe80::/56")
	flag.UintVar(&cnf.Lease.IPv6AB, "v6ab", cnf.Lease.IPv6AB, "Mask of IPv6 allocated block. Example: 64")
	flag.StringVar(&cnf.Lease.IPv4, "v4", cnf.Lease.IPv4, "Main IPv4 address pool. Example: 192.168.0.0/16")
	flag.UintVar(&cnf.Lease.IPv4AB, "v4ab", cnf.Lease.IPv4AB, "Mask of IPv4 allocated block. Example: 24")

	flag.Parse()
}

// Check - check config parametrs
func (cnf *Config) Check() error {
	// no config for starting server
	if !cnf.Server.UnixSocket && cnf.Server.Address == "" {
		return errors.New("Select one of server configuration: UNIX socket or TCP Address and port. Bouth can't be empty")
	}

	// no ip alocated blocks and no leases filename (or file must be Wipe)
	if cnf.Lease.IPv6 == "" && cnf.Lease.IPv4 == "" && (cnf.Lease.File == "" || cnf.Lease.Wipe) {
		return errors.New("No leases configuration")
	}

	// check for file exist
	if cnf.Lease.File != "" && !cnf.Lease.Wipe {
		if lf, err := os.Stat(cnf.Lease.File); os.IsNotExist(err) || lf.IsDir() || err != nil {
			return errors.New(cnf.Lease.File + " - lease file not exist or not have permissions")
		}
	}

	return nil
}

// ConfigHelp - print flag defaults
func ConfigHelp() {
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

	case bool:
		if r, err := strconv.ParseBool(env); err == nil {
			return r
		}
		return def
	}

	return def
}
