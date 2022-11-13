let gpiop = require('rpi-gpio').promise;

const trace = false

const gpioPin = 7;
let sLastSendToggleValue = 0;

/* some definitions from the IRremote Arduino Library */
const rc5AddressBits = 5;
const rc5CommandBits = 6;
const rc5ExtBits = 6;
const rc5CommandFieldBit = 1;
const rc5ToggleBit = 1;
const rc5Bits =
  rc5CommandFieldBit + rc5ToggleBit + rc5AddressBits + rc5CommandBits; // 13
const RC5XBits = rc5Bits + rc5ExtBits; // 19
const rc5Unit = 889; // (32 cycles of 36 kHz)
const rc5Duration = 15 * rc5Unit; // 13335
const rc5RepeatPeriod = 128 * rc5Unit; // 113792
const rc5RepeatSpace = rc5RepeatPeriod - rc5Duration; // 100 ms

/*
 *  normal Philips RC-5 as in the https://en.wikipedia.org/wiki/RC-5
 *  code taken from IRremote with some changes
 */

async function sendRC5 (aAddress, aCommand, aNumberOfRepeats) {
  gpiop.write(gpioPin, false);

  let tIRData = (aAddress & 0x1f) << rc5CommandBits;

  if (aCommand < 0x40) {
    // set field bit to lower field / set inverted upper command bit
    tIRData |= 1 << (rc5ToggleBit + rc5AddressBits + rc5CommandBits);
  } else {
    // let field bit zero
    aCommand &= 0x3f;
  }

  tIRData |= aCommand;

  tIRData |= 1 << rc5Bits;

  if (sLastSendToggleValue == 0) {
    sLastSendToggleValue = 1;
    // set toggled bit
    tIRData |= 1 << (rc5AddressBits + rc5CommandBits);
  } else {
    sLastSendToggleValue = 0;
  }

  let tNumberOfCommands = aNumberOfRepeats + 1;

  while (tNumberOfCommands > 0) {
    for (let i = 13; 0 <= i; i--) {
      if (trace) {
        console.log((tIRData & (1 << i)) ? '1' : '0');
      }

      if (tIRData & (1 << i)) {
        await send_1()
      } else {
        await send_0()
      }

      if (trace) {
        console.log("");
      }
    }
    tNumberOfCommands--;
    if (tNumberOfCommands > 0) {
      // send repeated command in a fixed raster
      await delayMicroseconds(rc5RepeatSpace / 1000 / 1000);
    }
  }

  gpiop.write(gpioPin, false);
  return 0;
}

/*
 *  Marantz 20 bit RC5 extension, see
 *  http://lirc.10951.n7.nabble.com/Marantz-RC5-22-bits-Extend-Data-Word-possible-with-lircd-conf-semantic-td9784.html
 *  could be combined with sendRC5, but ATM split to simplify debugging
 */

async function sendrc5X (aAddress, aCommand, aExt, aNumberOfRepeats) {
  let tIRData = uint32_t(aAddress & 0x1f) << (rc5CommandBits + rc5ExtBits);

  gpiop.write(gpioPin, false);

  if (aCommand < 0x40) {
    // set field bit to lower field / set inverted upper command bit
    tIRData |=
      1 << (rc5ToggleBit + rc5AddressBits + rc5CommandBits + rc5ExtBits);
  } else {
    // let field bit zero
    aCommand &= 0x3f;
  }

  tIRData |= aExt & 0x3f;
  tIRData |= aCommand << rc5ExtBits;
  tIRData |= 1 << RC5XBITS;

  if (sLastSendToggleValue == 0) {
    sLastSendToggleValue = 1;
    // set toggled bit
    tIRData |= 1 << (rc5AddressBits + rc5CommandBits + rc5ExtBits);
  } else {
    sLastSendToggleValue = 0;
  }

  let tNumberOfCommands = aNumberOfRepeats + 1;

  while (tNumberOfCommands > 0) {
    for (let i = 19; 0 <= i; i--) {
      if (trace) {
        console.log((tIRData & (1 << i)) ? '1' : '0')
      }
      if (tIRData & (1 << i)) {
        await send_1()
      } else {
        await send_0()
      }
      if (i == 12) {
        if (trace) {
          console.log('<p>')
        }
        // space marker for marantz rc5 extension
        await delayMicroseconds(rc5Unit * 2 * 2);
      }
      if (trace) {
        console.log('')
      }
    }
    tNumberOfCommands--;
    if (tNumberOfCommands > 0) {
      // send repeated command in a fixed raster
      await delayMicroseconds(rc5Repeat_space / 1000 / 1000)
    }
  }
  gpiop.write(gpioPin, false);
  return 0;
}

async function send_0 () {
  gpiop.write(gpioPin, true);
  await delayMicroseconds(rc5Unit);
  gpiop.write(gpioPin, false);
  await delayMicroseconds(rc5Unit);
}

async function send_1 () {
  gpiop.write(gpioPin, false);
  await delayMicroseconds(rc5Unit);
  gpiop.write(gpioPin, true);
  await delayMicroseconds(rc5Unit);
}

function delayMicroseconds (ms) {
  return new Promise(resolve => setTimeout(resolve, ms));
}

///////////////

async function run () {
  gpiop
    .setup(gpioPin, gpiop.DIR_OUT)
    .then(async () => {
      gpiop.write(gpioPin, false);
      // gpiop.write(gpioPin, true);

      await delayMicroseconds(1000)

      console.log('turn on')
      await sendRC5(16, 12, 1);
      console.log('done')

      await delayMicroseconds(1000)

      console.log('bye bye')

      gpiop.write(gpioPin, false);
      return gpiop.destroy();
    })
    .catch((err) => {
      console.log('Error: ', err.toString());
    });
}

run()

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
