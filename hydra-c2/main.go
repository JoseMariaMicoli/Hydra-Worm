package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Heartbeat represents the data structure expected from the Rust Agent
type Heartbeat struct {
	AgentID   string  `json:"agent_id"`
	Transport string  `json:"transport"`
	Status    string  `json:"status"`
	Lambda    float64 `json:"lambda"`
}

func displaySplash() {
	// Using a simpler, cleaner ASCII format to avoid parser issues with backslashes
	banner := `
   __              __             
  / /_  __  ______/ /________ _   
 / __ \/ / / / __  / ___/ __ '/   
/ / / / /_/ / /_/ / /  / /_/ /    
/_/ /_/\__, / .___/\__,_/_/   \__,_/     
      /____/_/                        
  W O R M  -  O R C H E S T R A T O R`

	fmt.Println(banner)
	fmt.Println("\n      [ Phase 1.4 - Malleable Orchestrator Active ]")
	fmt.Printf("      [ Listening on :8080 | Time: %s ]\n\n", time.Now().Format(time.RFC822))
}

func main() {
	displaySplash()

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// Heartbeat Endpoint with Malleable Profile Inspection
	r.POST("/api/v1/heartbeat", func(c *gin.Context) {
		// --- Phase 1.4: Profile Inspection ---
		uAgent := c.GetHeader("User-Agent")
		hydraKey := c.GetHeader("X-Hydra-Key")

		var hb Heartbeat
		if err := c.ShouldBindJSON(&hb); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid telemetry packet"})
			return
		}

		// Log detailed telemetry including the malleable profile metadata
		fmt.Printf("[%s] HEARTBEAT | Agent: %-15s | Transport: %-18s | λ: %.2f\n",
			time.Now().Format("15:04:05"),
			hb.AgentID,
			hb.Transport,
			hb.Lambda,
		)
		
		if hydraKey != "" {
			fmt.Printf("      └─ Profile Match: [UA: %s] [Key: %s]\n", uAgent, hydraKey)
		}

		// Respond with a simple 200 OK and tasking
		c.JSON(http.StatusOK, gin.H{
			"status": "acknowledged",
			"task":   "SLEEP",
			"epoch":  time.Now().Unix(),
		})
	})

	r.Run(":8080")
}