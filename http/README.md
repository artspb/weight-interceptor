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
4. Stores intercepted body data.

### building

To build the service, the minimal version of Go 1.13 is required.

`go build -o intercetor weight-interceptor-http`

### running

* the binary from the previous step: `./intercept`
* the sources: `go run weight-inteceptor-http`

## Configuration

In a single-user environment, no special configuration is required. If several people are using a scale, it might make sense to add `data/users.txt` that's intended to provide a mapping between users and their weight. The file format is simple. Each line should contain a username and their weight range separated by a space. For instance, `jack 80 100`. Such file can be created even for a single user. If it's absent, the name `default` is used.

## Data processors

### Raw data

The processor stores a raw request that contains weight to `data/data.txt`. It can serve as a backup. Also, it doesn't split data between users. A post-processing is required to extract timestamp and weight.

### CSV

The processor creates a CSV file for each user. It can be easily read by humans without additional manipulations with data. The CSV files are located under `data` and named after users, e.g., `jack.csv`.

### Google Fit

An additional configuration is required to start using this data processor. First, one need to create an app and register an OAuth client according to [this instruction](https://developers.google.com/fit/rest/v1/get-started). The only required permission is `fitness.body.write`. The credentials have to be placed under `data/credentials.json` (the file can be downloaded from [Google Console](https://console.cloud.google.com/apis/credentials)). Second, one has to authenticate each user, e.g., `interceptor fit jack`. The procedure is interactive and requires opening a URL and pasting a token. The token is then stored in a JSON file, e.g., `data/jack.json`.

The processor sends weight to a corresponding Google Account. If the procedure fails, it stores the data and performs a retry next time the weight arrives. Additionally, the retry is performed on the application start.

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

The request data consist of three parts. The first part consists of four digits, e.g., `01b8`. The first pair seems to be a descending sequence. The second pair is either `b8` or rarely `f8` in my case. The part's purpose remains unclear. The second part is the time of measurement e.g. `13262d3c` (eight digits). The third part is the data itself e.g. `1a68` (four digits). Example: `24BBBBBBBBBBBBSSSSSSSSSSSS01b813262d3c1a680000NN000000000000CCCCCCCC`.

The response data is always `01000000`. That means that the response is constant: `A0 0000000000000001000000000000000000000000000000bec650a1`.

###  22 Termination request

The request data is always `0000000000000000`, that's Bridge ID and CRC32 that differ. Example: `22BBBBBBBBBBBB0000000000000000000000000000000000000000000000CCCCCCCC`.

The response data is always `01000000`. That means that the response is constant: `A20000000000000000000000000000000000000000000000c9950d3f`.

### 29 Termination

The request data is always `0000000000000000`, that's Bridge ID and CRC32 that differ. Example: `22BBBBBBBBBBBB0000000000000000000000000000000000000000000000CCCCCCCC`.

There's no response as an HTTP connection gets closed.
