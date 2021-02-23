package bltmqtt

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"fmt"
	"crypto/tls"
	"crypto/x509"
	"time"
	"strings"
	"strconv"
	"log"
	"io/ioutil"
	"math/rand"
	"os"
	"net/url"
	"math"
	types "github.com/stronnag/bbl2kml/pkg/api/types"
	geo "github.com/stronnag/bbl2kml/pkg/geo"
	options "github.com/stronnag/bbl2kml/pkg/options"
	mission "github.com/stronnag/bbl2kml/pkg/mission"
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

func NewTlsConfig(cafile string) (*tls.Config, string) {
	if len(cafile) == 0 {
		return nil, "tcp"
	} else {
		certpool := x509.NewCertPool()
		ca, err := ioutil.ReadFile(cafile)
		if err != nil {
			log.Fatalln(err.Error())
		}
		certpool.AppendCertsFromPEM(ca)
		return &tls.Config{
			RootCAs:            certpool,
			InsecureSkipVerify: true, ClientAuth: tls.NoClientCert,
		},
			"ssl"
	}
}

func NewMQTTClient() *MQTTClient {
	var broker string
	var topic string
	var port int
	var cafile string
	var user string
	var passwd string

	rand.Seed(time.Now().UnixNano())

	if options.Mqttopts == "" {
		return nil
	}

	u, err := url.Parse(options.Mqttopts)
	if err != nil {
		log.Fatal(err)
	}

	broker = u.Hostname()
	port, _ = strconv.Atoi(u.Port())

	if len(u.Path) > 0 {
		topic = u.Path[1:]
	}

	up := u.User
	user = up.Username()
	passwd, _ = up.Password()

	q := u.Query()
	ca := q["cafile"]
	if len(ca) > 0 {
		cafile = ca[0]
	}
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

	tlsconf, scheme := NewTlsConfig(cafile)
	if u.Scheme == "ws" {
		scheme = "ws"
	}
	if u.Scheme == "wss" {
		tlsconf = &tls.Config{RootCAs: nil, ClientAuth: tls.NoClientCert}
		scheme = "wss"
	}

	if tlsconf == nil && (u.Scheme == "mqtts" || u.Scheme == "ssl") {
		tlsconf = &tls.Config{RootCAs: nil, ClientAuth: tls.NoClientCert}
		scheme = "ssl"
	}

	if len(os.Getenv("NOVERIFYSSL")) > 0 && tlsconf != nil {
		tlsconf.InsecureSkipVerify = true
	}

	clientid := fmt.Sprintf("%x", rand.Int63())
	opts := mqtt.NewClientOptions()

	mpath := ""
	if scheme == "ws" || scheme == "wss" {
		mpath = "/mqtt"
	}
	hpath := fmt.Sprintf("%s://%s:%d%s", scheme, broker, port, mpath)

	opts.AddBroker(hpath)
	opts.SetTLSConfig(tlsconf)
	opts.SetClientID(clientid)
	opts.SetUsername(user)
	opts.SetPassword(passwd)
	opts.SetDefaultPublishHandler(messagePubHandler)

	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}
	return &MQTTClient{client: client, topic: topic}
}

func (m *MQTTClient) publish(msg string) {
	token := m.client.Publish(m.topic, 0, false, msg)
	token.Wait()
}

/* Test brokers
   test.mosquitto.org 1883, 8883 8080, 8081 (ws)
   broker.hivemq.com  1883, 8000 (ws)
   broker.emqx.io    1883, 8883, 8083, 8084 (ws)
*/

func make_bullet_msg(b types.LogItem, homeamsl float64, elapsed int, ncells int, tgt int, nvs int) string {
	var sb strings.Builder

	sb.WriteString("flt:")
	sb.WriteString(strconv.Itoa(elapsed))
	sb.WriteByte(',')
	elapsed += 60
	sb.WriteString("ont:")
	sb.WriteString(strconv.Itoa(elapsed))
	sb.WriteByte(',')

	sb.WriteString("ran:")
	sb.WriteString(strconv.Itoa(int(b.Roll * 10)))
	sb.WriteByte(',')

	sb.WriteString("pan:")
	sb.WriteString(strconv.Itoa(int(b.Pitch * 10)))
	sb.WriteByte(',')

	sb.WriteString("hea:")
	sb.WriteString(strconv.Itoa(int(b.Cse)))
	sb.WriteByte(',')

	sb.WriteString("ggc:")
	sb.WriteString(strconv.Itoa(int(b.Cog)))
	sb.WriteByte(',')

	sb.WriteString("alt:")
	sb.WriteString(strconv.Itoa(int(b.Alt * 100)))
	sb.WriteByte(',')

	sb.WriteString("asl:")
	sb.WriteString(strconv.Itoa(int(b.GAlt)))
	sb.WriteByte(',')

	sb.WriteString("gsp:")
	sb.WriteString(strconv.Itoa(int(b.Spd * 100)))
	sb.WriteByte(',')

	sb.WriteString("bpv:")
	if options.Bulletvers == 2 {
		sb.WriteString(strconv.Itoa(int(b.Volts * 100)))
	} else {
		sb.WriteString(fmt.Sprintf("%.2f", float64(b.Volts)))
	}
	sb.WriteByte(',')

	avc := b.Volts / float64(ncells)
	sb.WriteString("acv:")
	if options.Bulletvers == 2 {
		sb.WriteString(strconv.Itoa(int(avc * 100)))
	} else {
		sb.WriteString(fmt.Sprintf("%.2f", avc))
	}
	sb.WriteByte(',')

	sb.WriteString("cad:")
	if options.Bulletvers == 2 {
		sb.WriteString(strconv.Itoa(int(b.Energy)))
	} else {
		sb.WriteString(fmt.Sprintf("%.0f", b.Energy))
	}
	sb.WriteByte(',')

	sb.WriteString("cud:")
	if options.Bulletvers == 2 {
		sb.WriteString(strconv.Itoa(int(b.Amps * 100)))
	} else {
		sb.WriteString(fmt.Sprintf("%.2f", b.Amps))
	}
	sb.WriteByte(',')

	//	rssi := 100 * int(b.Rssi) / 255
	sb.WriteString("rsi:")
	sb.WriteString(strconv.Itoa(int(b.Rssi)))
	sb.WriteByte(',')

	sb.WriteString("gla:")
	if options.Bulletvers == 2 {
		sb.WriteString(strconv.Itoa(int(b.Lat * 10000000)))
	} else {
		sb.WriteString(fmt.Sprintf("%.8f", b.Lat))
	}
	sb.WriteByte(',')

	sb.WriteString("glo:")
	if options.Bulletvers == 2 {
		sb.WriteString(strconv.Itoa(int(b.Lon * 10000000)))
	} else {
		sb.WriteString(fmt.Sprintf("%.8f", b.Lon))
	}
	sb.WriteByte(',')

	sb.WriteString("gsc:")
	sb.WriteString(strconv.Itoa(int(b.Numsat)))
	sb.WriteByte(',')

	sb.WriteString("ghp:")
	if options.Bulletvers == 2 {
		sb.WriteString(strconv.Itoa(int(b.Hdop)))
	} else {
		hdop := float64(b.Hdop) / 100.0
		sb.WriteString(fmt.Sprintf("%.1f", hdop))
	}
	sb.WriteByte(',')

	sb.WriteString("3df:")
	if b.Fix != 0 {
		sb.WriteString("1")
	} else {
		sb.WriteString("0")
	}
	sb.WriteByte(',')

	sb.WriteString("hds:")
	sb.WriteString(strconv.Itoa(int(b.Vrange)))
	sb.WriteByte(',')

	bearing := (b.Bearing + 180) % 360
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

	if tgt != 0 {
		sb.WriteString(fmt.Sprintf("cwn:%d,nvs:%d,", tgt, nvs))
	}

	armed := b.Status & 1
	sb.WriteString(fmt.Sprintf("arm:%d", armed))

	return sb.String()
}

func make_bullet_home(hlat float64, hlon float64, halt float64) string {
	var sb strings.Builder
	sb.WriteString("cs:JRandomUAV,")
	sb.WriteString("hla:")
	if options.Bulletvers == 2 {
		sb.WriteString(strconv.Itoa(int(hlat * 10000000)))
	} else {
		sb.WriteString(fmt.Sprintf("%.8f", hlat))
	}
	sb.WriteByte(',')
	sb.WriteString("hlo:")
	if options.Bulletvers == 2 {
		sb.WriteString(strconv.Itoa(int(hlon * 10000000)))
	} else {
		sb.WriteString(fmt.Sprintf("%.8f", hlon))
	}
	sb.WriteByte(',')
	sb.WriteString("hal:")
	if options.Bulletvers == 2 {
		sb.WriteString(strconv.Itoa(int(halt * 100)))
	} else {
		sb.WriteString(fmt.Sprintf("%.0f", halt))
	}
	return sb.String()
}

func make_bullet_mode(mode string, ncells int, hwfail bool) string {
	var sb strings.Builder
	if ncells > 0 {
		sb.WriteString("bcc:")
		sb.WriteString(strconv.Itoa(ncells))
		sb.WriteByte(',')
	}

	sb.WriteString("ftm:")
	sb.WriteString(mode)
	hwok := 1
	if hwfail {
		hwok = 0
	}
	sb.WriteString(fmt.Sprintf(",css:3,hwh:%d,", hwok))
	return sb.String()
}

func get_cells(vbat float64) int {
	ncell := 0
	for i := 1; i < 10; i++ {
		vmin := 3.0 * float64(i)
		vmax := 4.22 * float64(i)
		if vbat < vmax && vbat > vmin {
			ncell = i
			break
		}
	}
	return ncell
}

func output_message(c *MQTTClient, wfh *os.File, msg string, et time.Time) {
	if c != nil {
		c.publish(msg)
	}
	if wfh != nil {
		lt := et.UnixNano() / 1000000
		fmt.Fprintf(wfh, "%d|%s\n", lt, msg)
	}
}

func MQTTGen(s types.LogSegment) {
	ncells := 0
	var wfh *os.File
	tgt := 0
	nvs := 0

	c := NewMQTTClient()
	var err error
	if options.Outdir != "" {
		wfh, err = os.Create(options.Outdir)
		if err == nil {
			defer wfh.Close()
		}
	}

	if wfh == nil && c == nil {
		log.Fatal("Need at least a broker or log file")
	}

	var lastm time.Time
	laststat := uint8(255)
	fmode := ""
	mstrs := []string{}
	var ms *mission.Mission
	wps := ""
	if len(options.Mission) > 0 {
		var err error
		_, ms, err = mission.Read_Mission_File(options.Mission)
		if err == nil {
			var sb strings.Builder
			for k, mi := range ms.MissionItems {
				if geo.Getfrobnication() && mi.Is_GeoPoint() {
					ms.MissionItems[k].Lat, ms.MissionItems[k].Lon, _ = geo.Frobnicate_move(ms.MissionItems[k].Lat, ms.MissionItems[k].Lon, 0)
				}
				act, ok := mission.ActionMap[mi.Action]
				if !ok {
					act = 1
				}
				sb.Reset()
				sb.WriteString(fmt.Sprintf("wpno:%d,la:%d,lo:%d,al:%d,ac:%d,", mi.No,
					int(10000000*ms.MissionItems[k].Lat), int(10000000*ms.MissionItems[k].Lon),
					mi.Alt*100, act))
				if mi.P1 != 0 {
					sb.WriteString(fmt.Sprintf("p1:%d,", mi.P1))
				}
				if mi.P2 != 0 {
					sb.WriteString(fmt.Sprintf("p2:%d,", mi.P2))
				}
				if mi.P3 != 0 {
					sb.WriteString(fmt.Sprintf("p3:%d,", mi.P3))
				}
				sb.WriteString(fmt.Sprintf("el:%d,", int32(s.H.HomeAlt)+mi.Alt))
				if k == len(ms.MissionItems)-1 {
					sb.WriteString("f:165")
				}
				mstrs = append(mstrs, sb.String())
			}
			wps = fmt.Sprintf("wpc:%d,wpv:1,", len(ms.MissionItems))
		} else {
			fmt.Fprintf(os.Stderr, "* Failed to read mission file %s\n", options.Mission)
		}
	}

	miscout := 10
	// ensure once / minute or one two minues foe low prio data
	if options.Intvl > 6000 {
		miscout = 60 * 1000 / options.Intvl
		if miscout < 1 {
			miscout = 1
		}
	}

	var st time.Time
	for i, b := range s.L.Items {
		if i == 0 {
			st = b.Utc
			output_message(c, wfh, "Connected to flmqtt - pseudo/bullet/log/generator", b.Utc)
			output_message(c, wfh, "wpc:0,wpv:0,flt:0,ont:60", b.Utc)
		}

		et := int(b.Utc.Sub(st).Seconds())
		stat := b.Status >> 2

		if ncells == 0 {
			ncells = get_cells(b.Volts)
		}

		if b.Fmode != laststat {
			if laststat == types.FM_WP {
				tgt = 0
				nvs = 0
				output_message(c, wfh, "cwn:0,nvs:0", b.Utc)
			}
			if options.Bulletvers == 2 {
				switch b.Fmode {
				case types.FM_MANUAL:
					fmode = "1"
				case types.FM_ANGLE:
					fmode = "9"
				case types.FM_HORIZON:
					fmode = "10"
				case types.FM_ACRO:
					fmode = "11"
				case types.FM_AH:
					fmode = "8"
				case types.FM_PH:
					fmode = "4"
				case types.FM_WP:
					fmode = "7"
					if ms != nil {
						tgt = 1
						nvs = 5
					}
				case types.FM_RTH:
					fmode = "2"
				case types.FM_CRUISE3D:
					fmode = "5"
				case types.FM_LAUNCH:
					fmode = "9"
				default:
					fmode = "11"
				}
			} else {
				switch b.Fmode {
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
				if stat != 0 {
					fmode = "!FS!"
				}
			}
			laststat = b.Fmode
			msg := make_bullet_mode(fmode, ncells, b.HWfail)
			output_message(c, wfh, msg, b.Utc)
		}

		if i%miscout == 0 {
			msg := make_bullet_mode(fmode, ncells, b.HWfail)
			output_message(c, wfh, msg, b.Utc)
			msg = make_bullet_home(s.H.HomeLat, s.H.HomeLon, s.H.HomeAlt)
			output_message(c, wfh, msg, b.Utc)
			if len(mstrs) > 0 && i%2*miscout == 0 {
				for _, str := range mstrs {
					output_message(c, wfh, str, b.Utc)
				}
				output_message(c, wfh, wps, b.Utc)
			}
		}

		if b.Fmode == types.FM_WP && ms != nil {
			cdist := 1.50 * b.Spd * float64(options.Intvl/1000.0) / 1852.0
			for k, mi := range ms.MissionItems {
				if mi.Is_GeoPoint() {
					cse, d := geo.Csedist(b.Lat, b.Lon, mi.Lat, mi.Lon)
					if d < cdist {
						relb := math.Abs(cse - float64(b.Cse))
						/* fmt.Fprintf(os.Stderr, "Around WP %d brg=%.0f cse=%d d=%.1f (%.f) [%.1f]\n",
						mi.No, c, b.Cse, d*1852, relb, cdist*1852)*/
						if relb > 90 {
							if mi.No >= tgt { // may not have start of mission ....
								if k < len(ms.MissionItems)-1 {
									tgt += 1
									if ms.MissionItems[k+1].Action == "JUMP" {
										tgt = int(ms.MissionItems[k+1].P1)
									}
									if ms.MissionItems[k+1].Action == "RTH" {
										nvs = 4
									} else {
										nvs = 5
									}
									/* fmt.Fprintf(os.Stderr, "New target WP %d %d (%s)\n", tgt, k,
									ms.MissionItems[k+1].Action) */
								} else {
									tgt = 0
									nvs = 0
									output_message(c, wfh, "cwn:0,nvs:0", b.Utc)
								}
								break
							}
						}
					}
				}
			}
		}

		msg := make_bullet_msg(b, s.H.HomeAlt, et, ncells, tgt, nvs)
		output_message(c, wfh, msg, b.Utc)
		if c != nil && !lastm.IsZero() {
			tdiff := b.Utc.Sub(lastm)
			time.Sleep(tdiff)
		}
		lastm = b.Utc
	}
	// bizarrely, BulletGCSS expects the log to be "\n" line endings, apart from the last one
	if wfh != nil {
		fi, err := wfh.Stat()
		if err == nil {
			wfh.Truncate(fi.Size() - 1)
		}
	}
}
