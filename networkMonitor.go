package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// NetworkMetrics represents the core metrics we want to track
type NetworkMetrics struct {
	PacketLossRate    float64
	PacketDelay       float64
	GBR               bool // Guaranteed bits per second
	ConnTo5G		  bool
	LTECategory       int8 // LTE/5G
	
	// Application usage flags
	IoT 			  bool
	AVRGaming		  bool
	Healthcare        bool
	Industry40        bool
	IoTDevices        bool
	PublicSafety      bool
	SmartCityHome     bool
	SmartTransport    bool
	Smartphone        bool
	
	Timestamp         int8
	mutex             sync.RWMutex
}

// PacketInfo contains information about captured packets
type PacketInfo struct {
	Timestamp    time.Time
	Size         int
	Protocol     string
	SrcIP        net.IP
	DstIP        net.IP
	SrcPort      int
	DstPort      int
}

// NetworkMonitor is the main monitoring structure
type NetworkMonitor struct {
	Interface        string
	Handle          *pcap.Handle
	PacketChannel   chan PacketInfo
	Metrics         *NetworkMetrics
	PacketsSent     uint64
	PacketsReceived uint64
	PacketTimes     map[string]time.Time
	Running         bool
	mutex           sync.RWMutex
}

// Struct to hold all the settings
type Settings struct {
	GBR					bool `json:"GBR"`
	ConnTo5G			bool  `json:"ConnTo5G"`
	MinutesBeforeCloud	int8 `json:"minutesBeforeCloud"`
	IoT 			    bool  `json:"IoT"` 
	AVRGaming		    bool  `json:"AVRGaming"`
	Healthcare          bool  `json:"Healthcare"`
	Industry40          bool  `json:"Industry40"`
	IoTDevices          bool  `json:"IoTDevices"`
	PublicSafety        bool  `json:"PublicSafety"`
	SmartCityHome       bool  `json:"SmartCityHome"`
	SmartTransport      bool  `json:"SmartTransport"`
	Smartphone          bool  `json:"Smartphone"`
	LTECategory         int8  `json:"LTECategory"`
}

type ContextPair struct {
	Context *context.Context
	Cancel  *context.CancelFunc
	NM      *NetworkMonitor
}

var ContextManager []ContextPair
const TIME_FOR_DEVICE_VERIFICATION = 5

// NewNetworkMonitor creates a new network monitor instance
func NewNetworkMonitor(interfaceName string) (*NetworkMonitor, error) {
	// Open device for packet capture
	handle, err := pcap.OpenLive(interfaceName, 1600, true, pcap.BlockForever)
	if err != nil {
		return nil, fmt.Errorf("failed to open device %s: %v", interfaceName, err)
	}

	monitor := &NetworkMonitor{
		Interface:     interfaceName,
		Handle:        handle,
		PacketChannel: make(chan PacketInfo, 1000),
		Metrics: &NetworkMetrics{
			Timestamp:   int8(time.Now().Hour()),
		},
		PacketTimes: make(map[string]time.Time),
		Running:     false,
	}

	return monitor, nil
}

func (nm *NetworkMonitor) SetMetricsAsSettings(settings Settings){
	nm.Metrics.GBR = settings.GBR
	nm.Metrics.IoT = settings.IoT
	nm.Metrics.IoTDevices = settings.IoTDevices
	nm.Metrics.SmartCityHome = settings.SmartCityHome
	nm.Metrics.SmartTransport = settings.SmartTransport
	nm.Metrics.SmartCityHome = settings.SmartCityHome
	nm.Metrics.LTECategory = settings.LTECategory
	nm.Metrics.Healthcare = settings.Healthcare
	nm.Metrics.Industry40 = settings.Industry40
	nm.Metrics.AVRGaming = settings.AVRGaming
	nm.Metrics.ConnTo5G = settings.ConnTo5G
}

// Start begins the packet capture and analysis
func (nm *NetworkMonitor) Start(adaptor_no int8, settings *Settings, ctx context.Context, cancel context.CancelFunc, a *App) error {
	nm.mutex.Lock()
	nm.Running = true
	nm.mutex.Unlock()

	// Start packet capture goroutine
	go nm.capturePackets()
	
	// Start metrics analysis goroutine
	go nm.analyzeMetrics()
	
	// Start periodic reporting
	go nm.periodicReport(*settings, ctx, a)

	// Configure Settings
	nm.SetMetricsAsSettings(*settings)

	return nil
}

// Stop halts the monitoring
func (nm *NetworkMonitor) Stop() {
	nm.mutex.Lock()
	nm.Running = false
	nm.mutex.Unlock()
	
	nm.Handle.Close()
	close(nm.PacketChannel)
}

// capturePackets captures packets from the network interface
func (nm *NetworkMonitor) capturePackets() {
	packetSource := gopacket.NewPacketSource(nm.Handle, nm.Handle.LinkType())
	
	for packet := range packetSource.Packets() {
		nm.mutex.RLock()
		if !nm.Running {
			nm.mutex.RUnlock()
			break
		}
		nm.mutex.RUnlock()

		packetInfo := nm.extractPacketInfo(packet)
		
		select {
		case nm.PacketChannel <- packetInfo:
			nm.mutex.Lock()
			nm.PacketsReceived++
			nm.mutex.Unlock()
		default:
			// Channel full, drop packet
		}
	}
}

// extractPacketInfo extracts relevant information from a packet
func (nm *NetworkMonitor) extractPacketInfo(packet gopacket.Packet) PacketInfo {
	info := PacketInfo{
		Timestamp: time.Now(),
		Size:      len(packet.Data()),
		Protocol:  "Unknown",
	}

	// Extract network layer information
	if networkLayer := packet.NetworkLayer(); networkLayer != nil {
		switch networkLayer.LayerType() {
		case layers.LayerTypeIPv4:
			ipv4, _ := networkLayer.(*layers.IPv4)
			info.SrcIP = ipv4.SrcIP
			info.DstIP = ipv4.DstIP
			info.Protocol = "IPv4"
		case layers.LayerTypeIPv6:
			ipv6, _ := networkLayer.(*layers.IPv6)
			info.SrcIP = ipv6.SrcIP
			info.DstIP = ipv6.DstIP
			info.Protocol = "IPv6"
		}
	}

	// Extract transport layer information
	if transportLayer := packet.TransportLayer(); transportLayer != nil {
		switch transportLayer.LayerType() {
		case layers.LayerTypeTCP:
			tcp, _ := transportLayer.(*layers.TCP)
			info.SrcPort = int(tcp.SrcPort)
			info.DstPort = int(tcp.DstPort)
			info.Protocol += "/TCP"
			
		case layers.LayerTypeUDP:
			udp, _ := transportLayer.(*layers.UDP)
			info.SrcPort = int(udp.SrcPort)
			info.DstPort = int(udp.DstPort)
			info.Protocol += "/UDP"
			
		}
	}

	return info
}

// analyzeMetrics processes packets and updates metrics
func (nm *NetworkMonitor) analyzeMetrics() {
	for packetInfo := range nm.PacketChannel {
		nm.updatePacketLossRate()
		nm.updatePacketDelay(packetInfo)
	}
}

// updatePacketLossRate calculates packet loss rate
func (nm *NetworkMonitor) updatePacketLossRate() {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()

	if nm.PacketsSent > 0 {
		lossRate := float64(nm.PacketsSent-nm.PacketsReceived) / float64(nm.PacketsSent)
		nm.Metrics.PacketLossRate = lossRate
	}
}

// updatePacketDelay estimates packet delay (simplified)
func (nm *NetworkMonitor) updatePacketDelay(packetInfo PacketInfo) {
	/*
	// I am just taking processing time for now. 
	// Its impossible to calculated packetdelay throught only client. 
	// Server clock must be in sync too.
	// Later clever techniques like RTT / 2 can be applied.
	nm.Metrics.mutex.Lock()
	defer nm.Metrics.mutex.Unlock()
	
	// Simulate delay calculation
	// Well, thats the max we can do right now
	nm.Metrics.PacketDelay = float64(time.Since(packetInfo.Timestamp).Milliseconds())
	*/

	// Implementation failed
	nm.Metrics.mutex.Lock()
	defer nm.Metrics.mutex.Unlock()
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
    if err != nil {
		// fmt.Printf("failed to listen for ICMP packets: %v", err)
        return
    }
    defer conn.Close()

    // Create ICMP Echo Request
    msg := icmp.Message{
        Type: ipv4.ICMPTypeEcho,
        Code: 0,
        Body: &icmp.Echo{
            ID:   os.Getpid() & 0xffff,
            Seq:  1,
            Data: []byte("ping"),
        },
    }
    msgBytes, err := msg.Marshal(nil)
    if err != nil {
		// fmt.Printf("failed to marshal ICMP message: %v", err)
        return
    }

    dst, err := net.ResolveIPAddr("ip4", string(packetInfo.SrcIP))
    if err != nil {
		// fmt.Printf("failed to resolve IP: %v", err)
        return
    }

    start := time.Now()

    // Send Echo Request
    if _, err := conn.WriteTo(msgBytes, dst); err != nil {
		// fmt.Printf("failed to send ICMP request: %v", err)
        return
    }

    // Wait for Echo Reply
    reply := make([]byte, 1500)
    err = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
    if err != nil {
		// fmt.Printf("failed to set deadline: %v", err)
        return
    }

    n, _, err := conn.ReadFrom(reply)
    if err != nil {
		// fmt.Printf("failed to receive ICMP reply: %v", err)
        return 
	}

    duration := time.Since(start)
	nm.Metrics.PacketDelay = float64(duration.Milliseconds())

    parsedReply, err := icmp.ParseMessage(1, reply[:n])
    if err != nil {
		// fmt.Printf("failed to parse ICMP message: %v", err)
        return 
	}

    if parsedReply.Type == ipv4.ICMPTypeEchoReply {
        fmt.Printf("RTT to %s: %v\n", packetInfo.SrcIP, duration)
    } else {
		// fmt.Printf("got unexpected ICMP message type: %v", parsedReply.Type)
        return
    }


}

// periodicReport prints metrics periodically
func (nm *NetworkMonitor) periodicReport(settings Settings, ctx context.Context, a *App) {
	ticker := time.NewTicker(time.Duration(settings.MinutesBeforeCloud) * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			nm.printMetrics(a)
			SendJSONRequestToLambda(nm.Metrics, a)

			// Refreshing after every loop
			nm.mutex.Lock()
			nm.PacketsSent = 0
			nm.PacketsReceived = 0
			nm.mutex.Unlock()
		default:
			nm.mutex.RLock()
			if !nm.Running {
				nm.mutex.RUnlock()
				return
			}
			nm.mutex.RUnlock()
		
		}
	}
}

// printMetrics displays current metrics
func (nm *NetworkMonitor) printMetrics(a *App) {
	nm.Metrics.mutex.RLock()
	defer nm.Metrics.mutex.RUnlock()

	var appendableString string = fmt.Sprintf(`
		<b> Report from %s </b><br />
		Timestamp: %s<br />
		LTE Category: %d<br />
		Packet Loss Rate: %.2f<br />
		Packets Received: %d<br />
		GBR: %t<br />
		Packet Delay: %v<br />
		`, nm.Interface,
		time.Now().Format("15:04:05"),
		nm.Metrics.LTECategory,
		nm.Metrics.PacketLossRate,
		nm.PacketsReceived,
		nm.Metrics.GBR,
		nm.Metrics.PacketDelay,
	)

	a.showJS(false, appendableString)

	// Update
	// fmt.Printf("\n=== Network Monitoring Report from %s ===\n", nm.Interface)
	// fmt.Printf("Timestamp: %d\n", nm.Metrics.Timestamp)
	// fmt.Printf("LTE Category: %d\n", nm.Metrics.LTECategory)
	// fmt.Printf("Packet Loss Rate: %.2f\n", nm.Metrics.PacketLossRate)
	// fmt.Printf("GBR: %t\n", nm.Metrics.GBR)
	// fmt.Printf("Packet Delay: %v\n", nm.Metrics.PacketDelay)
	
	// fmt.Printf("\nApplication Usage:\n")
	// fmt.Printf("  AVR Gaming: %t\n", nm.Metrics.AVRGaming)
	// fmt.Printf("  Healthcare: %t\n", nm.Metrics.Healthcare)
	// fmt.Printf("  Industry 4.0: %t\n", nm.Metrics.Industry40)
	// fmt.Printf("  IoT Devices: %t\n", nm.Metrics.IoTDevices)
	// fmt.Printf("  IoT: %t\n", nm.Metrics.IoT)
	// fmt.Printf("  Public Safety: %t\n", nm.Metrics.PublicSafety)
	// fmt.Printf("  Smart City/Home: %t\n", nm.Metrics.SmartCityHome)
	// fmt.Printf("  Smart Transport: %t\n", nm.Metrics.SmartTransport)
	// fmt.Printf("  Smartphone: %t\n", nm.Metrics.Smartphone)
	
	// nm.mutex.RLock()
	// fmt.Printf("\nPackets Received: %d\n", nm.PacketsReceived)
	// nm.mutex.RUnlock()
	// fmt.Printf("==================================\n")
}

// func filterUselessDevices()
func filterUselessDevices(monitor *NetworkMonitor, ctx *(context.Context), cancel *(context.CancelFunc)){
	time.Sleep(TIME_FOR_DEVICE_VERIFICATION * time.Second)
	monitor.mutex.RLock()
	defer monitor.mutex.RUnlock()
	if monitor.PacketsSent == 0 && monitor.PacketsReceived == 0{
		// fmt.Println("No traffic found in this adapter. Stopping...")
		monitor.Stop()
		return
	}

	var newPair ContextPair
	newPair.Context = ctx
	newPair.Cancel = cancel
	newPair.NM = monitor
	ContextManager = append(ContextManager, newPair)
}

func (a *App) showJS(interruptProceeding bool, message string){
	if interruptProceeding {
		runtime.WindowExecJS(a.ctx, fmt.Sprintf("InteruptingError(`%s`);", message))
	}else{
		runtime.WindowExecJS(a.ctx, fmt.Sprintf("appendCard(`%s`);", message))
	}
}

func StartMonitoring(a *App) {

	// Importing settings
	settings, er := loadSettings(a)
	if er != nil {
		a.showJS(true, "Error loading settings")
	}
	
	devices, err := pcap.FindAllDevs()
	if err != nil {
		a.showJS(true, "Error finding devices")
	}
	
	if len(devices) == 0 {
		a.showJS(true, "No Network Devices found")
	}

	// Initialize monitor
	for i := 0; i < len(devices); i++ {
		monitor, err := NewNetworkMonitor(devices[i].Name)
		if err != nil {
			// log.Fatal(err)
			continue
		}
	    ctx, cancel := context.WithCancel(context.Background())
		defer monitor.Stop()
		// Start monitoring
		err = monitor.Start(int8(i+1), &settings, ctx, cancel, a)
		if err != nil {
			log.Fatal(err)
		}
		go filterUselessDevices(monitor, &ctx, &cancel)
	}

	for{
		select{
		case <-a.ctx.Done():
			os.Exit(0)
		default:		
			time.Sleep(1 * time.Second)
		}
	}
}