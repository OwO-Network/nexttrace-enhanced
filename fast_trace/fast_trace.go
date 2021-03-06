package fastTrace

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/xgadget-lab/nexttrace/config"
	"github.com/xgadget-lab/nexttrace/ipgeo"
	"github.com/xgadget-lab/nexttrace/printer"
	"github.com/xgadget-lab/nexttrace/reporter"
	"github.com/xgadget-lab/nexttrace/trace"
	"github.com/xgadget-lab/nexttrace/wshandle"
)

type FastTracer struct {
	Preference       config.Preference
	TracerouteMethod trace.Method
}

func (f *FastTracer) tracert(location string, ispCollection ISPCollection) {
	fmt.Printf("\033[1;33mã%s %s ã\033[0m\n", location, ispCollection.ISPName)
	fmt.Printf("traceroute to %s, 30 hops max, 32 byte packets\n", ispCollection.IP)
	ip := net.ParseIP(ispCollection.IP)
	var conf = trace.Config{
		BeginHop:         1,
		DestIP:           ip,
		DestPort:         80,
		MaxHops:          30,
		NumMeasurements:  3,
		ParallelRequests: 18,
		RDns:             !f.Preference.NoRDNS,
		IPGeoSource:      ipgeo.GetSource(f.Preference.DataOrigin),
		Timeout:          1 * time.Second,
	}

	if !f.Preference.TablePrintDefault {
		conf.RealtimePrinter = printer.RealtimePrinter
	}

	res, err := trace.Traceroute(f.TracerouteMethod, conf)

	if err != nil {
		log.Fatal(err)
	}

	if f.Preference.TablePrintDefault {
		printer.TracerouteTablePrinter(res)
		<-time.After(time.Second * 3)
	}

	println()

	if f.Preference.AlwaysRoutePath {
		r := reporter.New(res, ip.String())
		r.Print()
	}
}

func initialize() *FastTracer {
	configData, err := config.Read()

	// Initialize Default Config
	if err != nil || configData.DataOrigin == "" {
		if configData, err = config.AutoGenerate(); err != nil {
			log.Fatal(err)
		}
	}

	// Set Token from Config
	ipgeo.SetToken(configData.Token)

	return &FastTracer{
		Preference: configData.Preference,
	}
}

func (f *FastTracer) testAll() {
	f.testCT()
	println()
	f.testCU()
	println()
	f.testCM()
	println()
	f.testEDU()
}

func (f *FastTracer) testCT() {
	f.tracert(TestIPsCollection.Beijing.Location, TestIPsCollection.Beijing.CT163)
	f.tracert(TestIPsCollection.Shanghai.Location, TestIPsCollection.Shanghai.CT163)
	f.tracert(TestIPsCollection.Shanghai.Location, TestIPsCollection.Shanghai.CTCN2)
	f.tracert(TestIPsCollection.Hangzhou.Location, TestIPsCollection.Hangzhou.CT163)
	f.tracert(TestIPsCollection.Guangzhou.Location, TestIPsCollection.Guangzhou.CT163)
}

func (f *FastTracer) testCU() {
	f.tracert(TestIPsCollection.Beijing.Location, TestIPsCollection.Beijing.CU169)
	f.tracert(TestIPsCollection.Shanghai.Location, TestIPsCollection.Shanghai.CU169)
	f.tracert(TestIPsCollection.Shanghai.Location, TestIPsCollection.Shanghai.CU9929)
	f.tracert(TestIPsCollection.Hangzhou.Location, TestIPsCollection.Hangzhou.CU169)
	f.tracert(TestIPsCollection.Guangzhou.Location, TestIPsCollection.Guangzhou.CU169)
}

func (f *FastTracer) testCM() {
	f.tracert(TestIPsCollection.Beijing.Location, TestIPsCollection.Beijing.CM)
	f.tracert(TestIPsCollection.Shanghai.Location, TestIPsCollection.Shanghai.CM)
	f.tracert(TestIPsCollection.Hangzhou.Location, TestIPsCollection.Hangzhou.CM)
	f.tracert(TestIPsCollection.Guangzhou.Location, TestIPsCollection.Guangzhou.CM)
}

func (f *FastTracer) testEDU() {
	f.tracert(TestIPsCollection.Beijing.Location, TestIPsCollection.Beijing.EDU)
	f.tracert(TestIPsCollection.Shanghai.Location, TestIPsCollection.Shanghai.EDU)
	f.tracert(TestIPsCollection.Hangzhou.Location, TestIPsCollection.Hangzhou.EDU)
	f.tracert(TestIPsCollection.Hefei.Location, TestIPsCollection.Hefei.EDU)
	// ç§æç½ææ¶ç®å¨EDUéé¢ï¼ç­æ¿å°äºè¶³å¤å¤çæ°æ®ååç¦»åºå»ï¼åç¬ç¨äºæµè¯
	f.tracert(TestIPsCollection.Hefei.Location, TestIPsCollection.Hefei.CST)
	f.tracert(TestIPsCollection.Changsha.Location, TestIPsCollection.Changsha.EDU)
}

func FastTest(tm bool) {
	var c string

	fmt.Println("æ¨æ³æµè¯åªäºISPçè·¯ç±ï¼\n1. å½ååç½\n2. çµä¿¡\n3. èé\n4. ç§»å¨\n5. æè²ç½")
	fmt.Print("è¯·éæ©éé¡¹ï¼")
	fmt.Scanln(&c)

	ft := initialize()

	if !tm {
		ft.TracerouteMethod = trace.ICMPTrace
		fmt.Println("æ¨å°é»è®¤ä½¿ç¨ICMPåè®®è¿è¡è·¯ç±è·è¸ªï¼å¦ææ¨æ³ä½¿ç¨TCP SYNè¿è¡è·¯ç±è·è¸ªï¼å¯ä»¥å å¥ -T åæ°")
	} else {
		ft.TracerouteMethod = trace.TCPTrace
	}

	if strings.ToUpper(ft.Preference.DataOrigin) == "LEOMOEAPI" {
		// å»ºç« WebSocket è¿æ¥
		w := wshandle.New()
		w.Interrupt = make(chan os.Signal, 1)
		signal.Notify(w.Interrupt, os.Interrupt)
		defer func() {
			w.Conn.Close()
		}()
	}

	switch c {
	case "1":
		ft.testAll()
	case "2":
		ft.testCT()
	case "3":
		ft.testCU()
	case "4":
		ft.testCM()
	case "5":
		ft.testEDU()
	default:
		ft.testAll()
	}
}
