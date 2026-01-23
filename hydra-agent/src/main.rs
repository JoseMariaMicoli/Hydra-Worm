// Project: Hydra-Worm Agent
// Phase: 1.2 - Temporal Evasion (Poisson Jitter)

use std::thread;
use std::time::Duration;
use rand_distr::{Distribution, Exp};
use rand::thread_rng;

/// The core contract for all communication modules.
trait Transport {
    fn send_heartbeat(&self) -> Result<(), &'static str>;
    fn get_name(&self) -> String;
}

// --- TRANSPORT TIER 1: CLOUD API ---
struct CloudTransport { endpoint: String }
impl Transport for CloudTransport {
    fn send_heartbeat(&self) -> Result<(), &'static str> {
        println!("[!] C2-CLOUD: Attempting connection to {}...", self.endpoint);
        Err("403 Forbidden") 
    }
    fn get_name(&self) -> String { "Cloud-API".into() }
}

// --- TRANSPORT TIER 2: MALLEABLE DIRECT ---
struct MalleableTransport { c2_ip: String }
impl Transport for MalleableTransport {
    fn send_heartbeat(&self) -> Result<(), &'static str> {
        println!("[!] C2-DIRECT: Mimicking HTTPS to {}...", self.c2_ip);
        Err("TCP RST")
    }
    fn get_name(&self) -> String { "Malleable-Direct".into() }
}

// --- TRANSPORT TIER 3: P2P MESH ---
struct P2PTransport;
impl Transport for P2PTransport {
    fn send_heartbeat(&self) -> Result<(), &'static str> {
        println!("[!] C2-P2P: Broadcasting to local mesh...");
        Ok(())
    }
    fn get_name(&self) -> String { "P2P-Gossip-Mesh".into() }
}

// --- THE ENGINE ---
struct Agent {
    transport: Box<dyn Transport>,
    failures: u32,
    state: u8,
    lambda: f64, // Average heartbeats per second
}

impl Agent {
    fn new() -> Self {
        Self {
            transport: Box::new(CloudTransport { endpoint: "graph.microsoft.com".into() }),
            failures: 0,
            state: 0,
            lambda: 0.2, // Default: ~1 heartbeat every 5 seconds
        }
    }

    /// Calculates the next stochastic sleep interval based on an Exponential Distribution
    fn get_next_sleep(&self) -> Duration {
        let exp = Exp::new(self.lambda).unwrap();
        let seconds = exp.sample(&mut thread_rng());
        
        // Safety bounds: Min 1s, Max 60s to keep the simulation controlled
        let capped_seconds = seconds.min(60.0).max(1.0);
        
        println!("[?] Jitter Engine: λ={:.2} | Next heartbeat in {:.2}s", self.lambda, capped_seconds);
        Duration::from_secs_f64(capped_seconds)
    }

    fn run(&mut self) {
        loop {
            println!("\n[*] Strategy: {}", self.transport.get_name());
            match self.transport.send_heartbeat() {
                Ok(_) => self.failures = 0,
                Err(e) => {
                    println!("[-] Failure: {}", e);
                    self.failures += 1;
                }
            }

            if self.failures >= 3 {
                self.mutate();
            }

            // Apply Temporal Evasion
            thread::sleep(self.get_next_sleep());
        }
    }

    fn mutate(&mut self) {
        self.state = (self.state + 1) % 3;
        self.failures = 0;
        println!("[!!!] ALERT: Triggering Transport Mutation...");

        // Adjust lambda based on state: be "slower" (stealthier) as we lose options
        self.lambda = match self.state {
            0 => 0.2,  // Cloud: Normal
            1 => 0.1,  // Direct: Slower
            _ => 0.05, // P2P: Very Slow (1 every 20s avg)
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
    println!("      [ Phase 1.2 - Temporal Evasion Active ]\n");
}

fn main() {
    display_splash();
    let mut agent = Agent::new();
    agent.run();
}