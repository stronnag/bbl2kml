package bltmqtt

import (
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"time"
	"strings"
	"strconv"
	"math/rand"
	"os"
	types "github.com/stronnag/bbl2kml/pkg/api/types"
)

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connect lost: %v\n", err)
}

type MQTTClient struct {
	client mqtt.Client
	topic  string
}

func NewMQTTClient(broker string, topic string, port int) *MQTTClient {

	if broker == "" {
		broker = "broker.emqx.io"
	}
	if topic == "" {
		topic = fmt.Sprintf("org/mwptools/mqtt/loglayer/_%x", rand.Int())
		fmt.Fprintf(os.Stderr, "using random topic \"%s\"", topic)
	}

	if port == 0 {
		port = 1883
	}
	clientid := fmt.Sprintf("mwp_%x", rand.Int())
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port))
	opts.SetClientID(clientid)
	opts.SetUsername("")
	opts.SetPassword("")
	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	return &MQTTClient{client: client, topic: topic}
}

func (m *MQTTClient) publish(msg string) {
	token := m.client.Publish(m.topic, 0, false, msg)
	token.Wait()
}

func (m *MQTTClient) sub() {
	token := m.client.Subscribe(m.topic, 1, nil)
	token.Wait()
	fmt.Printf("Subscribed to topic: %s\n", m.topic)
}


/* Test brokers
   mqtt.eclipse.org
   test.mosquitto.org
   broker.hivemq.com
   mqtt.flespi.io
   mqtt.dioty.co
   mqtt.fluux.io
   broker.emqx.io
*/

func make_bullet_msg(b types.LogItem, homeamsl float64, elapsed int) string {
	var sb strings.Builder

	sb.WriteString("flt:")
	sb.WriteString(strconv.Itoa(elapsed))
	sb.WriteByte(',')
	elapsed += 60
	sb.WriteString("ont:")
	sb.WriteString(strconv.Itoa(elapsed))
	sb.WriteByte(',')

	sb.WriteString("ran:")
	sb.WriteString(strconv.Itoa(int(b.Roll) * 10))
	sb.WriteByte(',')

	sb.WriteString("pan:")
	sb.WriteString(strconv.Itoa(int(b.Pitch) * 10))
	sb.WriteByte(',')

	sb.WriteString("hea:")
	sb.WriteString(strconv.Itoa(int(b.Cse)))
	sb.WriteByte(',')

	sb.WriteString("ggc:")
	sb.WriteString(strconv.Itoa(int(b.Cog)))
	sb.WriteByte(',')

	sb.WriteString("alt:")
	sb.WriteString(strconv.Itoa(int(b.Alt) * 100))
	sb.WriteByte(',')

	sb.WriteString("asl:")
	sb.WriteString(strconv.Itoa(int(b.GAlt)))
	sb.WriteByte(',')

	sb.WriteString("gsp:")
	sb.WriteString(strconv.Itoa(int(b.Spd) * 100))
	sb.WriteByte(',')

	sb.WriteString("bpv:")
	sb.WriteString(fmt.Sprintf("%.2f", float64(b.Volts)))
	sb.WriteByte(',')

	sb.WriteString("cad:")
	sb.WriteString(fmt.Sprintf("%.0f", b.Energy))
	sb.WriteByte(',')

	sb.WriteString("cud:")
	sb.WriteString(fmt.Sprintf("%.2f", b.Amps))
	sb.WriteByte(',')

	//	rssi := 100 * int(b.Rssi) / 255
	sb.WriteString("rsi:")
	sb.WriteString(strconv.Itoa(int(b.Rssi)))
	sb.WriteByte(',')

	sb.WriteString("gla:")
	sb.WriteString(fmt.Sprintf("%.8f", b.Lat))
	sb.WriteByte(',')

	sb.WriteString("glo:")
	sb.WriteString(fmt.Sprintf("%.8f", b.Lon))
	sb.WriteByte(',')

	sb.WriteString("gsc:")
	sb.WriteString(strconv.Itoa(int(b.Numsat)))
	sb.WriteByte(',')

	sb.WriteString("ghp:")
	hdop := float64(b.Hdop) / 100.0
	sb.WriteString(fmt.Sprintf("%.1f", hdop))
	sb.WriteByte(',')

	sb.WriteString("3df:")
	sb.WriteString(strconv.Itoa(int(b.Fix)))
	sb.WriteByte(',')

	sb.WriteString("hds:")
	sb.WriteString(strconv.Itoa(int(b.Vrange)))
	sb.WriteByte(',')

	bearing := 180 - b.Bearing
	if bearing < 0 {
		bearing += 360
	}

	sb.WriteString("hdr:")
	sb.WriteString(strconv.Itoa(int(bearing)))
	sb.WriteByte(',')

	sb.WriteString("trp:")
	sb.WriteString(strconv.Itoa(b.Throttle))
	sb.WriteByte(',')

	fs := (b.Status & 2) >> 1
	sb.WriteString("fs:")
	sb.WriteString(strconv.Itoa(int(fs)))
	sb.WriteByte(',')

	armed := b.Status & 1
	sb.WriteString(fmt.Sprintf("arm:%d", armed))
	return sb.String()
}

func make_bullet_home(hlat float64, hlon float64, halt float64) string {
	var sb strings.Builder
	sb.WriteString("cs:JRandomUAV,")
	sb.WriteString("hla:")
	sb.WriteString(fmt.Sprintf("%.8f", hlat))
	sb.WriteByte(',')

	sb.WriteString("hlo:")
	sb.WriteString(fmt.Sprintf("%.8f", hlon))
	sb.WriteByte(',')
	sb.WriteString("hal:")
	sb.WriteString(fmt.Sprintf("%.0f", halt))

	return sb.String()
}

func make_bullet_mode(mode string, ncells int) string {
	var sb strings.Builder
	if ncells > 0 {
		sb.WriteString("bcc:")
		sb.WriteString(strconv.Itoa(ncells))
		sb.WriteByte(',')
	}

	sb.WriteString("ftm:")
	sb.WriteString(mode)
	sb.WriteString(",css:1")
	return sb.String()
}

func get_cells(vbat float64) int {
	ncell := 0
	for i := 1; i < 10; i++ {
		v := 3.0 * float64(i)
		if vbat < v {
			ncell = i - 1
			break
		}
	}
	return ncell
}

func MQTTGen(broker string, topic string, port int, s types.LogSegment) {
	ncells := 0
	c := NewMQTTClient(broker, topic, port)
	var lastm time.Time

	laststat := uint8(0)
	fmode := ""

	st := time.Now()
	for i, b := range s.L.Items {
		now := time.Now()
		et := int(now.Sub(st).Seconds())
		stat := b.Status >> 2

		if ncells == 0 {
			ncells = get_cells(b.Volts)
		}

		if b.Fmode != laststat {
			switch stat {
			case types.FM_MANUAL:
				fmode = "MANU"
			case types.FM_ANGLE:
				fmode = "ANGL"
			case types.FM_HORIZON:
				fmode = "HOR"
			case types.FM_ACRO:
				fmode = "ACRO"
			case types.FM_AH:
				fmode = "A H"
			case types.FM_PH:
				fmode = "P H"
			case types.FM_WP:
				fmode = "WP"
			case types.FM_RTH:
				fmode = "RTH"
			case types.FM_CRUISE3D:
				fmode = "3CRS"
			case types.FM_LAUNCH:
				fmode = "LNCH"
			default:
				fmode = "ACRO"
			}
			laststat = stat
			msg := make_bullet_mode(fmode, ncells)
			c.publish(msg)
		}

		if i%10 == 0 {
			msg := make_bullet_mode(fmode, ncells)
			c.publish(msg)
			msg = make_bullet_home(s.H.HomeLat, s.H.HomeLon, s.H.HomeAlt)
			c.publish(msg)
		}

		msg := make_bullet_msg(b, s.H.HomeAlt, et)
		c.publish(msg)
		if !lastm.IsZero() {
			tdiff := b.Utc.Sub(lastm)
			time.Sleep(tdiff)
		}
		lastm = b.Utc
	}
}
