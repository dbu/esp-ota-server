#include <ESP8266WiFi.h>

// refresh frequency in milliseconds. needs to be lower than TTL on server side
#define REGISTRY_REFRESH 1800000

unsigned long nextUpdate = ULONG_MAX;

void loopRegisterIp(void) {
  if (nextUpdate > millis()) {
    return;
  }
  if (WiFi.status() != WL_CONNECTED) {
    return;
  }

  registerIp();
}

void registerIp(void) {
  WiFiClient client;
  HTTPClient http;

  http.begin(client, LICHTWURFEL_SERVER REGISTER_PATH);
  http.addHeader(F("x-esp8266-sta-mac"), WiFi.macAddress());
  char body[3+5+17+11+34];
  snprintf(body, sizeof(body), "{\"ip\":\"%s\",\"network\":\"%s\"}", WiFi.localIP().toString().c_str(), conf.ssid);
  int httpCode = http.POST(body);
  if (httpCode > 0) {
    Serial.println(F("Updated IP"));
  } else {
    Serial.print("Error code: ");
    Serial.println(httpCode);
  }

  http.end();

  nextUpdate = millis() + REGISTRY_REFRESH;
}
