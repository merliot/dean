![Gopher Thing](images/gopher_cloud.png)

Dean is a 100% Go framework.  It's currently under develoment.

Dean wraps a Server around a Thing.  The Server extends Go's net/http web server.
The Thing has a ServeHTTP() handler, subscribed message handlers, and a Run loop.

Dean also works with [TinyGo](https://tinygo.org")*.  You can deploy a Thing on a microcontroller using
TinyGo.

### Example

There is an example in examples/main.go.  To run:

```
go run examples/main.go
```

This will start a web server on port :8080.  Open a browser to http://localhost:8080.

To run on a microcontroller with TinyGo*:

```
tinygo flash -monitor -target nano-rp2004 -stack-size 4KB ~/work/dean/example/
```

This will start a web server on port :8080.  Open a browser to http://[IP addr]:8080, where [IP addr] is the microcontroller's IP address.

\* requires netdev PRs to TinyGo

