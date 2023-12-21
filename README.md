ESP-OTA-Server
==============

Forked from https://github.com/vooon/esp-ota-server

Very simple OTA firmware server suitable for built-in [ESP8266 HTTP Updater][1] and [ESP32 HTTP Updater][2].

Main purpose is to serve firmware files and passing MD5 hash -- to verify flashing.

Options:
- `-s` `--bind` listen address (default `:8092`)
- `-d` `--data-dir` data storage location. `<data-dir>/<project>/<file.bin>`

OTA URL: `http://<server-bind-host>/bin/<project>/<file.bin>`


TODO:
- Handle version
- Provide syncing data files when we don't want to overwrite the whole SPIFFS image

[1]: https://github.com/esp8266/Arduino/tree/master/libraries/ESP8266httpUpdate
[2]: https://github.com/suculent/esp32-http-update
