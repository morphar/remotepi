let gpiop = require('rpi-gpio').promise;

gpiop
  .setup(7, gpiop.DIR_OUT)
  .then(async () => {
    gpiop.write(7, false);
    gpiop.write(7, true);
    // gpiop.write(7, false);
    // return gpiop.destroy();
  })
  .catch((err) => {
    console.log('Error: ', err.toString());
  });

///////////////


//  void setup()   {
//    //start serial connection
//    Serial.begin(115200);
//    // configure output pin
//    pinMode(IR_TX_PIN, OUTPUT);
//    // Connect to Wi-Fi
//    WiFi.setHostname("marantz");

//    // Disable wifi power-save mode to improve stability
//    WiFi.mode (WIFI_STA);
//    esp_wifi_set_ps(WIFI_PS_NONE);

//    WiFi.begin(ssid, password, 0, bssid);
//    // or, if you dont want bssid locking, use
//    // WiFi.begin(ssid, password);

//    while (WiFi.status() != WL_CONNECTED) {
//      delay(1000);
//      Serial.println("Connecting to WiFi..");
//    }
//    // Print ESP Local IP Address
//    Serial.println(WiFi.localIP());

//    // Route for root / web page
//    server.on("/", HTTP_GET, [](AsyncWebServerRequest *request){
//      request->send_P(200, "text/html", index_html);
//    });

//    // Send a GET request to <ESP_IP>/update?button=<name>
//    server.on("/update", HTTP_GET, [] (AsyncWebServerRequest *request) {
//      String inputMessage1;
//      // GET input1 value on <ESP_IP>/update?output=<inputMessage1>&state=<inputMessage2>
//      if (request->hasParam("button")) {
//        inputMessage1 = request->getParam("button")->value();
//        if (strcmp ("standby", inputMessage1.c_str()) == 0) {
//           sendRC5(16, 12, 1);
//        }
//        if (strcmp ("phono", inputMessage1.c_str()) == 0) {
//           sendRC5(21, 63, 1);
//        }
//        if (strcmp ("cd", inputMessage1.c_str()) == 0) {
//           sendRC5(20, 63, 1);
//        }
//        if (strcmp ("tuner", inputMessage1.c_str()) == 0) {
//           sendRC5(17, 63, 1);
//        }
//        if (strcmp ("aux1", inputMessage1.c_str()) == 0) {
//           sendrc5X(16, 0, 6, 1);
//        }
//        if (strcmp ("aux2", inputMessage1.c_str()) == 0) {
//           sendrc5X(16, 0, 7, 1);
//        }
//        if (strcmp ("dcc", inputMessage1.c_str()) == 0) {
//           sendRC5(23, 63, 1);
//        }
//        if (strcmp ("tape", inputMessage1.c_str()) == 0) {
//           sendRC5(18, 63, 1);
//        }
//        if (strcmp ("volume_up", inputMessage1.c_str()) == 0) {
//           sendRC5(16, 16, 1);
//        }
//        if (strcmp ("volume_down", inputMessage1.c_str()) == 0) {
//           sendRC5(16, 17, 1);
//        }
//      }
//      else {
//        inputMessage1 = "No message sent";
//      }
//      Serial.print("Button: ");
//      Serial.print(inputMessage1);
//      Serial.print("\n");
//      request->send(200, "text/plain", "OK\n");
//    });

//    // Start server
//    server.begin();
//  }
