```markdown
           / /_  __  ______  __/ /__________ _   
          / __ \/ / / / __ \/ __  / ___/ __ `/   
         / / / / /_/ / /_/ / /_/ / /  / /_/ /    
        /_/ /_/\__, / .___/\__,_/_/   \__,_/     
   _      ____/____/_/___  ____ ___              
  | | /| / / __ \/ __ \/ __ `__ \                
  | |/ |/ / /_/ / /_/ / / / / / /                
  |__/|__/\____/_/ .__/_/ /_/ /_/                 
                /_/                              

      [ 2026 Offensive R&D Research Project ]
```
---

# HYDRA-WORM: THE GHOST ORCHESTRATOR

> **Project Phase:** 1.5 - Low-Level Failsafes & Malleable Integration
> **Research Status:** RED TEAM R&D / DEFENSIVE GAP ANALYSIS
> **Core Principle:** Multi-Tiered Transport Resilience & Temporal Evasion



## I. FULL PROJECT ROADMAP (SPRINTS & SUB-PHASES)

| Phase | Sub-Phase | Focus / Technical Deliverable | Status |
| --- | --- | --- | --- |
| **Sprint 1: Stealth** |  |  |  |
| [x] | 1.1 | **Transport Abstraction:** Rust Traits for hot-swappable communication modules. | DONE |
| [x] | 1.2 | **Temporal Evasion:** NHPP (Non-Homogeneous Poisson Process) jitter engine. | DONE |
| [x] | 1.3 | **C2 Bootstrap:** Go-based Gin/Echo backend with TLS 1.3 and mutual auth. | DONE |
| [ ] | 1.4 | **Malleable Profiles:** Dynamic HTTP/2 header and JA3/S fingerprint randomization. | **ACTIVE** |
| [ ] | 1.5 | **Failsafe Stack:** Integration of ICMP, NTP, and DNS transports. | **ACTIVE** |
| **Sprint 2: Recon** |  |  |  |
| [ ] | 2.1 | **Artifact Harvesting:** Parsing `known_hosts`, RDP `MRU`, and `bash_history`. | PLANNED |
| [ ] | 2.2 | **Environment Context:** IMDSv2 (AWS/Azure/GCP) and Container (K8s/Docker) detection. | PLANNED |
| [ ] | 2.3 | **EDR/XDR Fingerprinting:** Enumerating drivers and hooked APIs for evasion logic. | PLANNED |
| [ ] | 2.4 | **Structured Telemetry:** Protobuf-encoded reporting for minimal network signature. | PLANNED |
| **Sprint 3: Propagation** |  |  |  |
| [ ] | 3.1 | **Credential Management:** Secure handling and reuse of captured NTLM/Kerberos/SSH tokens. | PLANNED |
| [ ] | 3.2 | **P2P Discovery:** mDNS/UDP/LLMNR gossip mesh for internal peer discovery. | PLANNED |
| [ ] | 3.3 | **Infection Engine:** Multithreaded propagation via SMB, SSH, and WMI mocks. | PLANNED |
| [ ] | 3.4 | **Safety Throttle:** Propagation rate-limiting and global "Kill-Switch" broadcast. | PLANNED |
| **Sprint 4: DFIR** |  |  |  |
| [ ] | 4.1 | **LotL Persistence:** Implementation via WMI Event Subs, Systemd timers, and GPO. | PLANNED |
| [ ] | 4.2 | **Syscall Evasion:** Refactoring core logic for Direct/Indirect Syscalls (bypassing `ntdll`). | PLANNED |
| [ ] | 4.3 | **Atomic Destruction:** Self-deletion logic including secure file wiping. | PLANNED |
| [ ] | 4.4 | **CLI Completion:** Integration of help-autocomplete and shell commands. | PLANNED |

---

## II. RULES OF ENGAGEMENT (ROE)

1. **Authorization:** Execution is strictly prohibited on any system without explicit, written "Stay Out of Jail" documentation.
2. **Environment Isolation:** Simulation must be restricted to logically/physically segmented lab environments.
3. **Safety Throttle:** Propagation is limited to a maximum of 3 hops per 24 hours to prevent uncontrolled "Worm Storms."
4. **The Kill-Switch:** A global "Kill-Switch" broadcast (via UDP/5353) must be available at all times to force immediate self-deletion.
5. **Data Sovereignty:** No PII or sensitive corporate data exfiltration. Telemetry is restricted to status and non-sensitive context.

---

## III. THE 6-TIER MUTATION HIERARCHY

| Tier | Protocol | Stealth Method | Use Case |
| --- | --- | --- | --- |
| **1** | **Cloud-API** | Domain Fronting (Graph/S3) | Primary C2 (Highest Trust) |
| **2** | **Malleable** | HTTP/2 + JA3/S Rotation | Bypass TLS Fingerprinting |
| **3** | **P2P Mesh** | UDP mDNS / TCP Gossip | Lateral Movement / Air-gap |
| **4** | **ICMP** | Echo Request Payloads | Firewall Bypass (Ping allowed) |
| **5** | **NTP** | Transmit Timestamp Covert | High-Stealth / Low-Bandwidth |
| **6** | **DNS** | Base64 Subdomain Tunneling | Last-Resort / Locked-Down Segments |

---

## IV. TECHNICAL WHITE PAPER

### **1. Mathematical Foundation: NHPP Temporal Evasion**

Hydra-Worm utilizes a **Non-Homogeneous Poisson Process (NHPP)** to generate heartbeat intervals, mimicking stochastic user behavior.

The probability of  beacons in the interval  is given by:

Where the integrated intensity  is defined as:

**Intensity Factors ():**

* **System Noise:**  increases during high disk I/O to blend with background activity.
* **Office Hours:**  follows a sinusoidal curve to mimic standard 9-to-5 usage.

### **2. Malleable Profiles (Phase 1.4)**

* **JA3/S Randomization:** Using `rustls`, the agent dynamically alters the TLS *Client Hello*. By rotating Cipher Suites, Extensions, and Elliptic Curves, it produces unique fingerprints that evade database-led blocking.
* **HTTP/2 Header Rotation:** Dynamic injection of `User-Agent`, `Accept-Language`, and custom `X-Hydra-Key` headers. Headers are reordered per connection to mimic browser-specific behavior (Chrome, Firefox, Safari).

### **3. Low-Level Covert Channels (Phase 1.5)**

* **ICMP Tunneling:** Encapsulates XOR-encrypted telemetry within the 32-64 byte trailing payload of standard ICMP Type 8 (Echo Request).
* **NTP Clockwork:** Leverages the 64-bit **Transmit Timestamp**. The 32-bit "fractional" portion—typically jittery sub-second data—is replaced with encrypted C2 status codes.
* **DNS Failsafe:** Last-resort exfiltration via Base64-encoded subdomains (e.g., `telemetry.ns1.hydra.lab`) directed at the Orchestrator's nameserver.

### **4. Endpoint Stealth: Syscall Evasion & Zeroization**

* **Direct Syscalls:** Bypasses EDR hooks on `ntdll.dll` by dynamically resolving syscall numbers and executing them via inline assembly:
```asm
mov r10, rcx
mov eax, [syscall_number]
syscall

```


* **Memory Sanitization:** Implements the `zeroize` trait. Upon state transition, sensitive structs are overwritten using `volatile` operations to prevent forensic memory dumps from recovering keys or C2 metadata.

---

## V. MITRE ATT&CK® MAPPING

| Tactic | Technique ID | Name | Application |
| --- | --- | --- | --- |
| **Reconnaissance** | T1595 | Active Scanning | Passive ARP/mDNS sniffing. |
| **Defense Evasion** | T1573.002 | Asymmetric Encrypted Channel | Phase 1.4 Malleable HTTPS. |
| **Defense Evasion** | T1562.001 | Impair Defenses | Direct Syscall execution in Rust. |
| **Defense Evasion** | T1070 | Indicator Removal | Zeroizing RAM via `zeroize` crate. |
| **Command & Control** | T1071.004 | DNS Tunneling | Phase 1.5 DNS Failsafe. |
| **Command & Control** | T1095 | Non-Application Layer Protocol | Phase 1.5 ICMP Tunneling. |
| **Command & Control** | T1027 | Obfuscated Information | Phase 1.5 NTP Covert signaling. |

---

## VI. NIST SP 800-61 R3 INCIDENT RESPONSE TEMPLATE

### **1. Detection & Identification**

* **Signature:** Identify TLS handshakes where JA3 fingerprints do not match the declared `User-Agent`.
* **Heuristic:** Baseline standard NTP/ICMP volume. Flag packets where entropy in timestamp or payload fields exceeds  bits/byte.
* **Tooling:** Use Zeek/Suricata to analyze Inter-Arrival Time (IAT) for Poisson-distributed anomalies.

### **2. Containment & Eradication**

* **Containment:** Block outbound UDP/123 to non-standard IPs and sinkhole detected DNS subdomains. Disable mDNS (UDP/5353) to collapse the P2P mesh.
* **Eradication:** Perform a memory scan to identify process hollowing. Audit `~/.ssh/known_hosts` to identify at-risk lateral targets.

---

## VII. LEGAL DISCLAIMER

This software is for **authorized educational and research purposes only**. Use on systems without prior written consent is strictly prohibited and illegal. The developers assume no liability for misuse.

---