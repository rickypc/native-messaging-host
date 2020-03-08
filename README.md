[![Build](https://img.shields.io/travis/rickypc/native-messaging-host)](https://bit.ly/2ItWBWM)
[![Coverage](https://img.shields.io/codecov/c/github/rickypc/native-messaging-host)](https://bit.ly/2TwjOyb)
[![Dependabot](https://api.dependabot.com/badges/status?host=github&repo=rickypc/native-messaging-host)](https://bit.ly/2KIM5vs)
[![License](https://img.shields.io/github/license/rickypc/native-messaging-host)][8]

# Native Messaging Host Module for Go

native-messaging-host is a module for sending [native messaging protocol][1]
message marshalled from struct and receiving [native messaging protocol][1]
message unmarshalled to struct. native-messaging-host can auto-update itself
using update URL that response with Google Chrome [update manifest][2],
as well as it provides hook to install and uninstall manifest file to
[native messaging host location][3].

## Installation and Usage

Package documentation can be found on [GoDoc][4].

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
host.Uninstall()
```

Contributing
-
If you would like to contribute code to Native Messaging Host repository you can do so
through GitHub by forking the repository and sending a pull request.

If you do not agree to [Contribution Agreement](CONTRIBUTING.md), do not
contribute any code to Native Messaging Host repository.

When submitting code, please make every effort to follow existing conventions
and style in order to keep the code as readable as possible. Please also include
appropriate test cases.

That's it! Thank you for your contribution!

License
-
Copyright (c) 2018 - 2020 Richard Huang.

This utility is free software, licensed under: [Mozilla Public License (MPL-2.0)][8].

Documentation and other similar content are provided under [Creative Commons Attribution-NonCommercial-ShareAlike 4.0 International License][9].

[1]: https://bit.ly/3axo5Xv
[2]: https://bit.ly/2vOdAR5
[3]: https://bit.ly/2TuQrMw
[4]: https://bit.ly/2TMGqcj
[5]: https://bit.ly/2Tt4Poo
[6]: https://bit.ly/3cAVAdq
[7]: https://bit.ly/3aDA1Hv
[8]: https://mzl.la/2vLmCye
[9]: https://bit.ly/2SMCRlS
