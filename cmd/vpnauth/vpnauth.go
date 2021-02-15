package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/cryptohippie/webauthclient"
)

var (
	interfaceName string
	clientID      string
	password      string
	passwordFile  string
	timeoutFactor int
	sourceAddress string
)

func init() {
	flag.StringVar(&interfaceName, "interface", "", "network interface to make connections from")
	flag.StringVar(&clientID, "clientid", "", "account ClientID for authentication")
	flag.StringVar(&password, "password", "", "password for ClientID")
	flag.StringVar(&passwordFile, "passwordfile", "", "read password from file, first line is the password")
	flag.IntVar(&timeoutFactor, "timeoutfactor", 3, "factor to multiply the default timeout values with")
	flag.StringVar(&sourceAddress, "sourceaddr", "", "source address to bind to")
	flag.Parse()
}

func main() {
	var auth *webauthclient.Authenticator
	if interfaceName != "" && sourceAddress != "" {
		fmt.Fprint(os.Stderr, "Cannot set both -interface and -sourceaddr\n")
		os.Exit(1)
	}
	if interfaceName != "" {
		factory, err := webauthclient.ClientFactoryForInterface(interfaceName, uint64(timeoutFactor))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Interface error: %s\n", err)
			os.Exit(1)
		}
		auth = &webauthclient.Authenticator{
			HTTPClientFactory: factory,
		}
	}
	if sourceAddress != "" {
		sourceIP := net.ParseIP(sourceAddress)
		if sourceIP == nil {
			fmt.Fprintf(os.Stderr, "Given address is no valid IP address: %s\n", sourceAddress)
			os.Exit(1)
		}
		auth = &webauthclient.Authenticator{
			HTTPClientFactory: webauthclient.ClientFactoryForAddress(sourceIP, uint64(timeoutFactor)),
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
