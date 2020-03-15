# weight-interceptor-http

In March 2020, Soehnle has [terminated support](https://forms.office.com/Pages/ResponsePage.aspx?id=tzE8KfMtyECIyOIWNul_drKzR790OsJDlS5ggvmsdnZUOU1BTllLT1JSMlFPWVRBSTNLSEZVUVFOWC4u) of their [Web Connect scales](https://my.soehnle.com/). They plan to shut down their server by the end of March 2020. But scales already lost their ability to store and analyze body data. The goal of this project is to preserve this functionality even without a centralized service.

## Web Connect service description

Web Connect service consists of three components: Scale, Web Box, and an HTTP server. They have the following roles.

* Scale measures body data and sends it via a 433 MHz radio channel to Web Box.
* Web Box receives and temporary stores body data as well as sends it to the HTTP server.
* The HTTP server permanently stores body data and provides access to it for a web page and mobiles apps.

## Web Box description

Web Box is a bridge between a scale and Internet. It connects to a router via a LAN cable. IP address can be configured either via DHCP or set manually. It sends body data via unencrypted HTTP/1.1 to `bridge1.soehnle.de`. The later allows to intercept information and redirect it to a desired location. In order to do this, one need to have a device with a LAN port and perform the following steps using it.

1. Configure a DNS server to resolve `bridge1.soehnle.de` to a local IP address. One can use [this instruction](https://www.digitalocean.com/community/tutorials/how-to-configure-bind-as-a-private-network-dns-server-on-ubuntu-18-04) for Ubuntu 18.04 or find a similar one for the desired OS.
2. Configure a DHCP server to return the IP address of the DNS server from the previous step. [This article](
) can be useful for Ubuntu 18.04.
3. Run `weight-interceptor-http` on the local server to intercept traffic.

## weight-interceptor-http description

The service acts as a proxy between Web Box and the original HTTP server. It performs the following steps.

1. Finds data in the request.
2. Verifies CRC32 checksum.
3. Prepares its own response.
4. Finds the IP of the original server using an external DNS server.
5. Queries the original server.
6. Compares the server response with its own.
7. Returns either the server response or its own if the server doesn't respond.
8. Stores intercepted body data.

An external DNS server is used as it's expected that the local one resolves the target host to the local address. By default, Google Public DNS (`8.8.8.8`) is queried.

### building

To build the service, the minimal version of Go 1.13 is required.

`go build -o intercetor weight-interceptor-http`

### running

* the binary from the previous step: `./intercept`
* the sources: `go run weight-inteceptor-http`

## Communication protocol

Web Box sends a GET request in the following form every ten seconds.

`http://bridge1.soehnle.de/devicedataservice/dataservice?data=%request%`

The `%request%` part is a 68 symbols long string. Each symbol is a hex number of the following format.

`IIBBBBBBBBBBBBSSSSSSSSSSSSDDDDDDDDDDDDDDDD0000NN000000000000CCCCCCCC`

Code | Length | Description
--- | ---: | ---
I | 2 | Command
B | 12 | Bridge ID (written on the device)
S | 12 | (?) Scale ID, only with code 24
D | 16 | Data
0 | 4 | Always zero
N | 2 | (?), only with code 24
0 | 12 | Always zero
C | 8 | CRC32 with the IEEE 802.3 polynomial

Command can be one of the following.

Code | Description
--- | ---
25 | Sync request
28 | Sync
21 | Sync response
24 | Data transmission
22 | Termination request
29 | Termination

They can be divided in the following groups: sync, data, termination. Commands inside a group always go together. Each session consists of either a sync, a data, or both in the order sync-data always followed by the termination.

Here are possible combinations of codes: `25-28-21-22-29`, `24-22-29`, `25-28-21-24-22-29`. Web Box can repeat commands several times, but it doesn't happen to `29` as it closes an HTTP connection. Normally, when `24` is repeated it carries a new portion of body data.

The server responds with a `200 OK` message. Its body consists of another hex number 56 symbols long. The format is as follows.

`RR00000000000000DDDDDDDD000000000000000000000000CCCCCCCC`

Code | Length | Description
--- | ---: | ---
RR | 2 | Response code
0 | 14 | Always zero
D | 8 | Data
0 | 24 | Always zero
C | 8 | CRC32 with the IEEE 802.3 polynomial

Here are possible response codes.

Code | Description
--- | ---
A0 | Response to `25` or `24`
A5 | Response to `28`
A1 | Response to `21`
A2 | Response to `22`

`29` never gets a response. Below is an example of communication.

```
GET http://bridge1.soehnle.de/devicedataservice/dataservice?data=22XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX HTTP/1.1
Host: bridge1.soehnle.de
Connection: keep-alive


```
```
HTTP/1.1 200 OK
Date: Sun, 15 Mar 2020 16:20:31 GMT
Content-Length: 56
Content-Type: text/plain; charset=utf-8

A20000000000000000000000000000000000000000000000c9950d3f
```
```
GET http://bridge1.soehnle.de/devicedataservice/dataservice?data=29XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX HTTP/1.1
Host: bridge1.soehnle.de
Connection: close
```

The body always has an extra `\n` at the end that's why Go's HTTP client complains `Unsolicited response received on idle HTTP channel starting with "\n"; err=<nil>`.

### 25 Sync request

_Starts a sync sequence._

The request data is always `00000000413a4104`, that's Bridge ID and CRC32 that differ. Example: `25BBBBBBBBBBBB00000000000000000000413a4104000000000000000000CCCCCCCC`.

The response data is always `01000000`. That means that the response is constant: `A00000000000000001000000000000000000000000000000bec650a1`.

### 28 Sync

_Sends Web Box time._

The request data consists of two parts 8 symbols long each. The first part is always different, its purpose is unknown. The second part is hex digits that hold a number of seconds since 2010-01-01T00:00:00+01:00 e.g `13262c44`. Example: `28BBBBBBBBBBBB000000000000????????13262c44000000000000000000CCCCCCCC`.

The response data is always `01000000`. That means that the response is constant: `A5000000000000000100000000000000000000000000000056e5abd9`.

### 21 Sync response

_Requests server time._

The request data is always `0000000000000000`, that's Bridge ID and CRC32 that differ. Example: `21BBBBBBBBBBBB0000000000000000000000000000000000000000000000CCCCCCCC`.

The response data holds a server time in the same format. Example: `A10000000000000013262c580000000000000000000000004043c9fe`.

### 24 Data transmission

_Sends body data._

The request data consist of three parts. The first part is always `01b8`. It could be a type of data like weight, fat, and so on. The second part is the time of measurement e.g. `13262d3c`. The third part is the data itself e.g. `1a68`. Example: `24BBBBBBBBBBBBSSSSSSSSSSSS01b813262d3c1a680000NN000000000000CCCCCCCC`.

The response data is always `01000000`. That means that the response is constant: `A0 0000000000000001000000000000000000000000000000bec650a1`.

###  22 Termination request

The request data is always `0000000000000000`, that's Bridge ID and CRC32 that differ. Example: `22BBBBBBBBBBBB0000000000000000000000000000000000000000000000CCCCCCCC`.

The response data is always `01000000`. That means that the response is constant: `A20000000000000000000000000000000000000000000000c9950d3f`.

### 29 Termination

The request data is always `0000000000000000`, that's Bridge ID and CRC32 that differ. Example: `22BBBBBBBBBBBB0000000000000000000000000000000000000000000000CCCCCCCC`.

There's no response as an HTTP connection gets closed.
