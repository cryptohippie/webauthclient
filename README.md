# webauthclient
Programmatic interaction with CHAVPN authentication framework.

## How to use

1. Install go
2. Install client:\
   `go get -u github.com/cryptohippie/webauthclient/cmd/vpnauth`
3. Call authenticator:\
   `vpnauth -clientid YOUR-CLIENT-ID -password YOUR-PASSWORD`
4. Exit code 0, no output -> All went according to plan!

Alternatively you can use the binaries from vpnauth-bin/, they have been prebuilt for 
x86_64/amd64 systems.
