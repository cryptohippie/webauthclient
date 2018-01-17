package main

import (
	"bufio"
	"cryptohippie/webauthclient"
	"flag"
	"fmt"
	"os"
	"strings"
)

var (
	interfaceName string
	clientID      string
	password      string
	passwordFile  string
)

func init() {
	flag.StringVar(&interfaceName, "interface", "", "network interface to make connections from")
	flag.StringVar(&clientID, "clientid", "", "account ClientID for authentication")
	flag.StringVar(&password, "password", "", "password for ClientID")
	flag.StringVar(&passwordFile, "passwordfile", "", "read password from file, first line is the password")
	flag.Parse()
}

func main() {
	var auth *webauthclient.Authenticator
	if interfaceName != "" {
		factory, err := webauthclient.ClientFactoryForInterface(interfaceName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Interface error: %s\n", err)
			os.Exit(1)
		}
		auth = &webauthclient.Authenticator{
			HTTPClientFactory: factory,
		}
	}

	if passwordFile != "" {
		var f *os.File
		var err error
		switch passwordFile {
		case "-":
			f = os.Stdin
		default:
			f, err = os.Open(passwordFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "File error, cannot open: %s", passwordFile)
			}
			defer f.Close()
		}
		b := bufio.NewReader(f)
		p, err := b.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "File error, cannot read: %s", passwordFile)
		}
		password = strings.Trim(p, "\n\r")
	}

	err := auth.Authenticate(clientID, password)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Authentication error: %s\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}
