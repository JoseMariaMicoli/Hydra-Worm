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

To ensure your `README.md` renders perfectly on GitHub while maintaining the technical depth of the **6-Tier Architecture** and **Rules of Engagement (RoE)**, I have updated the file with the ````math` block syntax.

This version uses the file you provided as a base and integrates the full roadmap, the updated hierarchy (adding ICMP, NTP, and DNS), and the non-truncated mathematical formulas.

---

# HYDRA-WORM: THE GHOST ORCHESTRATOR

> **Project Phase:** 1.5 - Low-Level Failsafes & Malleable Integration
> **Research Status:** RED TEAM R&D / DEFENSIVE GAP ANALYSIS
> **Core Principle:** Multi-Tiered Transport Resilience & Temporal Evasion

```text
   __              __             
  / /_  __  ______/ /________ _   
 / __ \/ / / / __  / ___/ __ `/   
/ / / / /_/ / /_/ / /  / /_/ /    
/_/ /_/\__, / .___/\__,_/_/   \__,_/     
      /____/_/                        
  [ 2026 Offensive R&D Research Project ]
  [ Status: Phase 1.5 - Failsafe Integration ]

```

## I. PROJECT README & OPERATIONAL ROADMAP

### **Project Overview**

**Hydra-Worm** is a next-generation, research-oriented breach simulation framework. It utilizes a **Rust-based Agent** for low-level stealth and memory safety, and a **Go-based Orchestrator** for high-concurrency C2 operations. The framework simulates "Living off the Land" (LotL) propagation, polymorphic C2 evasion, and cross-platform lateral movement.

### **Legal Disclaimer & Rules of Engagement (RoE)**

> **CRITICAL LEGAL NOTICE:** This software is developed exclusively for **authorized Red Team Research, Development (R&D), and Defensive Gap Analysis**.
> 1. **Authorization:** Execution is strictly prohibited on any system without explicit, written "Stay Out of Jail" documentation.
> 2. **Environment Isolation:** The simulation must be restricted to logically or physically segmented lab environments.
> 3. **Resource Management:** Propagation and heartbeat intervals must be throttled to prevent Denial of Service (DoS) on network infrastructure.
> 4. **Cleanup Guarantee:** Every iteration must include a pre-verified, automated "Nuclear" cleanup routine.
> 5. **Safety Throttle:** Propagation is limited to a maximum of 3 hops per 24 hours to prevent uncontrolled "Worm Storms."
> 6. **The Kill-Switch:** A global "Kill-Switch" broadcast (via UDP/5353) must be available at all times to force immediate self-deletion.
> 
> 

### **Full Project Roadmap (Sprints & Sub-Phases)**

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
| [ ] | 4.4 | **CLI Completion:** Integration of help-autocomplete, shell commands, and README update. | PLANNED |

---

## II. THE 6-TIER MUTATION HIERARCHY

| Tier | Protocol | Stealth Method | Use Case |
| --- | --- | --- | --- |
| **1** | **Cloud-API** | Domain Fronting (Graph/S3) | Primary C2 (Highest Trust) |
| **2** | **Malleable** | HTTP/2 + JA3/S Rotation | Bypass TLS Fingerprinting |
| **3** | **P2P Mesh** | UDP mDNS / TCP Gossip | Lateral Movement / Air-gap |
| **4** | **ICMP** | Echo Request Payloads | Firewall Bypass (Ping allowed) |
| **5** | **NTP** | Transmit Timestamp Covert | High-Stealth / Low-Bandwidth |
| **6** | **DNS** | Base64 Subdomain Tunneling | Last-Resort / Locked-Down Segments |

---

## III. TECHNICAL WHITE PAPER: ADVANCED PERSISTENT WORM ARCHITECTURE

### **1. Mathematical Foundation: NHPP Temporal Evasion**

Hydra-Worm utilizes a **Non-Homogeneous Poisson Process (NHPP)** to generate heartbeat intervals.

The probability of  beacons in the interval  is given by:

```math
P[N(t+\tau)-N(t)=n] = \frac{[\Lambda(t,\tau)]^n}{n!} e^{-\Lambda(t,\tau)}

```

Where the integrated intensity  is defined as:

```math
\Lambda(t,\tau) = \int_{t}^{t+\tau} \lambda(s)ds

```

**Intensity Factors ():**

* **System Noise:**  increases during high disk I/O to blend with background activity.
* **Time of Day:**  follows a sinusoidal curve to mimic office hours.

### **2. Network Polymorphism: The Transport Abstraction Layer**

The agent manages a registry of **Transport Providers**. An internal `Decision Engine` monitors egress health; if a transport triggers an EDR alert or TCP Reset, the agent executes a "Hot Swap" of the interface.

* **Malleable Profiles (Phase 1.4):** Employs JA3/S fingerprint randomization via `rustls` and HTTP/2 header rotation (User-Agent, Accept-Language) to bypass DPI.
* **Covert Failsafes (Phase 1.5):** Includes ICMP Echo Request payload encapsulation and NTP Transmit Timestamp signaling.

### **3. Endpoint Stealth: Direct Syscalls & Memory Zeroization**

* **Syscall Evasion:** Bypasses `ntdll.dll` hooks by resolving syscall numbers from the disk-based version of `ntdll.dll` and executing them via assembly:

```asm
mov r10, rcx
mov eax, [syscall_number]
syscall

```

* **Memory Sanitization:** To thwart forensic dumps, the agent implements the `zeroize` trait. Sensitive structs are overwritten using `volatile` operations during state transitions.

---

## IV. MITRE ATT&CK® MAPPING (DETAILED)

| Tactic | Technique | ID | Hydra-Worm Implementation Detail |
| --- | --- | --- | --- |
| **Reconnaissance** | Active Scanning | T1595 | Passive ARP/mDNS sniffing to identify peers. |
| **Execution** | Shared Modules | T1129 | Polymorphic transport logic via reflective loading. |
| **Defense Evasion** | Direct System Calls | T1562.001 | Bypassing EDR shims via kernel syscalls in Rust. |
| **Defense Evasion** | Indicator Removal | T1070 | Zeroizing C2 metadata in RAM via `zeroize`. |
| **Command & Control** | Traffic Signaling | T1543 | **NHPP-based** heartbeats for stochastic behavior. |
| **Command & Control** | Multi-hop Proxy | T1090 | Routing traffic from restricted subnets through peers. |

---

## V. DFIR RESPONSE TEMPLATE (NIST SP 800-61 R3)

### **1. Detection (Preparation & Identification)**

* **Network:** Identify TLS handshakes with anomalous JA3/S fingerprints. Monitor for high-frequency UDP/5353 (mDNS) traffic.
* **Endpoint:** Monitor for the "Fork and Run" pattern or loading signed binaries into unusual processes.

### **2. Analysis (RS.AN)**

* **Memory:** Extract the PEB; look for evidence of **Reflective Loading**.
* **Traffic:** Use **Inter-Arrival Time (IAT)** analysis to identify the "lambda" fingerprint.

### **3. Containment & Eradication**

* **Containment:** Implement "Micro-segmentation" to break the mDNS/UDP peer discovery.
* **Eradication:** Verify removal of Scheduled Tasks and WMI Event Consumers. Check for "Phantom Tasks" in the registry.

---