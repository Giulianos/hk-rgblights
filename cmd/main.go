package main

// based on github.com/brutella/hc switch example

import (
    "log"
    "fmt"

    "github.com/lucasb-eyer/go-colorful"
    MQTT "github.com/eclipse/paho.mqtt.golang"
    "github.com/brutella/hc"
    "github.com/brutella/hc/accessory"
)

func publishColor(mqttClient MQTT.Client, deviceID string, h, s, v float64) {
    c := colorful.Hsv(h, s, v)
    log.Println(c)
    log.Printf("h, s, v: %f, %f, %f\n", h, s, v)
    r := int(c.R*1023)
    g := int(c.G*1023)
    b := int(c.B*1023)

    topic := fmt.Sprintf("device/%s", deviceID)
    message := fmt.Sprintf("%04d,%04d,%04d", r, g, b)
    mqttClient.Publish(topic, 0, false, message)
}

func main() {
    // connect to mqtt broker
    brokerAddr := "tcp://localhost:1883"
    opts := MQTT.NewClientOptions().AddBroker(brokerAddr)
    opts.SetClientID("hk_bridge")

    mqttClient := MQTT.NewClient(opts)
    if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
        panic(token.Error())
    }

    // current color
    var h, s, v float64

    // create an accessory
    acc := accessory.NewColoredLightbulb(accessory.Info{Name: "RGB Light"})
    acc.Lightbulb.On.OnValueRemoteUpdate(func(on bool) {
        if on {
            publishColor(mqttClient, "rgb_lights_001", h, s, v)
        } else {
            publishColor(mqttClient, "rgb_lights_001", h, s, 0)
        }
    })
    acc.Lightbulb.Brightness.OnValueRemoteUpdate(func(brightness int) {
        v = float64(brightness)/100
        publishColor(mqttClient, "rgb_lights_001", h, s, v)
        log.Println("Brightness was set to: ", brightness)
    })
    acc.Lightbulb.Saturation.OnValueRemoteUpdate(func(saturation float64) {
        s = saturation/100
        publishColor(mqttClient, "rgb_lights_001", h, s, v)
        log.Println("Saturation was set to: ", saturation)
    })
    acc.Lightbulb.Hue.OnValueRemoteUpdate(func(hue float64) {
        h = hue
        publishColor(mqttClient, "rgb_lights_001", h, s, v)
        log.Println("Hue was set to: ", hue)
    })

    // configure the ip transport
    config := hc.Config{Pin: "00102003"}
    t, err := hc.NewIPTransport(config, acc.Accessory)
    if err != nil {
        mqttClient.Disconnect(250)
        log.Panic(err)
    }

    hc.OnTermination(func() {
        <-t.Stop()
    })

    t.Start()
}
