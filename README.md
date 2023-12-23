Forked from https://github.com/vooon/esp-ota-server

ESP-OTA-Server
==============

Utility server to support ESP8266/ESP32 devices.

* Over the Air (OTA) firmware updates suitable for built-in [ESP8266 HTTP Updater][1] and [ESP32 HTTP Updater][2].
* Register the local network IP with the server - fallback for when mDNS does not work in your network.

Features
--------

### OTA

Handles requests for updated binary. The server checks the MD5 hash to decide if the binary needs to be updated or is
already at its latest version. The server also sends the `x-md5` header in the response to allow the built-in upgrader
to verify the binary.

The OTA URL looks like this: `http://<server>/bin/<project>/<file.bin>`

    curl -H "x-esp8266-mode: sketch" -H "x-esp8266-version: 1" localhost:8092/bin/mycoolproject/code.bin

### Register IP

The ESP does a web request to the server once the network is established. It needs to `POST` to `/register`, with
the `x-esp8266-sta-mac` header containing its MAC address (used to avoid duplicate entries) and a JSON body with the
`ip` and `network` properties.

    curl http://localhost:8092/register -H "x-esp8266-sta-mac: x123" -d '{"ip": "1.2.3.4", "network":"foo bar"}' -v

See [Example Sketch][doc/register-ip.ino] for an illustration how to register the IP.

Consumers can request the IP at `http://<server>/lookup/<network>`. If there is exactly one match, they are redirected,
if there are several, a page with the list of possible devices is rendered. If the network is unknown, a 404 Not Found
page is returned.

The network needs to be correctly written with special characters and whitespace, but ignores case for ease of use.

Startup
-------

Options:
- `-s` `--bind` listen address (default `:80`)
- `-d` `--data-dir` data storage location. Server will look for binaries at `<data-dir>/bin/<project>/<file.bin>`

Build
-----

To build the binary with your local operating system:

    make build-on-host

To build the binary for alpine - this will create a docker image esp-ota-server *and* update the espotad binary in the
root directory.
Note that we don't cache the vendors, so this is inefficient.

    make build-with-alpine

TODO
----
- Cache go vendors in Dockerfile build
- Handle version
- Provide syncing data files when we don't want to overwrite the whole SPIFFS image

[1]: https://github.com/esp8266/Arduino/tree/master/libraries/ESP8266httpUpdate
[2]: https://github.com/suculent/esp32-http-update
