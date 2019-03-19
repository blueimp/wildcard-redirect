# Wildcard Redirect
> HTTP service to redirect wildcard subdomains of an origin host to a new target
> host, e.g. `*.example.com` to `*.example.org`.

## Installation
The `wildcard-redirect` binary can be installed from source via
[go get](https://golang.org/cmd/go/):

```sh
go get github.com/blueimp/wildcard-redirect
```

## Usage
By default, `wildcard-redirect` listens on port `8080` on all interfaces and
redirects all origin subdomains to subdomains of the given target host:

```sh
wildcard-redirect example.org
```

### Options
Available options can be listed the following way:

```sh
wildcard-redirect --help
```

```
  -a string
    	TCP listen address (default ":8080")
  -q string
    	Original host query parameter (default "via")
  -s	Always redirect using HTTPS
```

By default, the original host domain is appended as query parameter:

```
http://test.example.org/?via=test.example.com
```

Setting the parameter name to the empty string disables appending the original
host domain:

```
wildcard-redirect -q '' example.org
```

If no target host is given, the `Location` header points to the original URL.

Using just the `-s` option, this can be used for a simple HTTP to HTTPS redirect
service:

```
wildcard-redirect -s
```

### Logging
Requests are logged in `JSON` format to `stdout`:

```json
{
  "Time": "2018-07-17T10:19:26.055298263Z",
  "RemoteIP": "::1",
  "Method": "GET",
  "Host": "test.example.com",
  "RequestURI": "/",
  "Referrer": "",
  "UserAgent": "curl/7.54.0",
  "ForwardedFor": "",
  "ForwardedProto": ""
}
```

## License
Released under the [MIT license](LICENSE.txt).
