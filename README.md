# Native Messaging Host Module for Go

native-messaging-host is a module for sending [native messaging protocol][1]
message marshalled from struct and receiving [native messaging protocol][1]
message unmarshalled to struct. native-messaging-host can auto-update itself
using update URL that response with Google Chrome [update manifest][2],
as well as it provides hook to install and uninstall manifest file to
[native messaging host location][3].

## Installation and Usage

Package documentation can be found on [GoDev][4].

Installation can be done with a normal `go get`:

```
$ go get github.com/rickypc/native-messaging-host
```

#### Receiving Message

```go
// Ensure func main returned after calling [runtime.Goexit][5].
defer os.Exit(0)

messaging := (&host.Host{}).Init()

// host.H is a shortcut to map[string]interface{}
request := &host.H{}

// Read message from os.Stdin to request.
if err := messaging.OnMessage(os.Stdin, request); err != nil {
  log.Fatalf("messaging.OnMessage error: %v", err)
}

// Log request.
log.Printf("request: %+v", request)
```

#### Sending Message

```go
messaging := (&host.Host{}).Init()

// host.H is a shortcut to map[string]interface{}
response := &host.H{"key":"value"}

// Write message from response to os.Stdout.
if err := messaging.PostMessage(os.Stdout, response); err != nil {
  log.Fatalf("messaging.PostMessage error: %v", err)
}

// Log response.
log.Printf("response: %+v", response)
```

#### Auto Update Configuration

updates.xml example for cross platform executable:

```xml
<?xml version='1.0' encoding='UTF-8'?>
<gupdate xmlns='http://www.google.com/update2/response' protocol='2.0'>
  <app appid='tld.domain.sub.app.name'>
    <updatecheck codebase='https://sub.domain.tld/app.download.all' version='1.0.0' />
  </app>
</gupdate>
```

updates.xml example for individual platform executable:

```xml
<?xml version='1.0' encoding='UTF-8'?>
<gupdate xmlns='http://www.google.com/update2/response' protocol='2.0'>
  <app appid='tld.domain.sub.app.name'>
    <updatecheck codebase='https://sub.domain.tld/app.download.darwin' os='darwin' version='1.0.0' />
    <updatecheck codebase='https://sub.domain.tld/app.download.linux' os='linux' version='1.0.0' />
    <updatecheck codebase='https://sub.domain.tld/app.download.exe' os='windows' version='1.0.0' />
  </app>
</gupdate>
```

```go
// It will do daily update check.
messaging := (&host.Host{
  AppName:   "tld.domain.sub.app.name",
  UpdateUrl: "https://sub.domain.tld/updates.xml", // It follows [update manifest][2]
  Version:   "1.0.0",                              // Current version, it must follow [SemVer][6]
}).Init()
```

#### Install and Uninstall Hooks

```go
// AllowedExts is a list of extensions that should have access to the native messaging host. 
// See [native messaging manifest][7]
messaging := (&host.Host{
  AppName:     "tld.domain.sub.app.name",
  AllowedExts: []string{"chrome-extension://XXX/", "chrome-extension://YYY/"},
}).Init()

...

// When you need to install.
if err := messaging.Install(); err != nil {
  log.Printf("install error: %v", err)
}

...

// When you need to uninstall.
if err := host.Uninstall(); err != nil {
  log.Printf("uninstall error: %v", err)
}
```

## Issues and Contributing

If you find an issue with this module, please report an issue. If you'd
like, we welcome any contributions. Fork this module and submit a pull
request.

[1]: https://bit.ly/3axo5Xv
[2]: https://bit.ly/2vOdAR5
[3]: https://bit.ly/2TuQrMw
[4]: https://bit.ly/2Tw22L6
[5]: https://bit.ly/2Tt4Poo
[6]: https://bit.ly/3cAVAdq
[7]: https://bit.ly/3aDA1Hv
