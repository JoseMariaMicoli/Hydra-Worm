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

# HYDRA-WORM: THE GHOST ORCHESTRATOR

> **Project Phase:** Artifact Harvesting: Parsing `known_hosts`, RDP `MRU`, and `bash_history`.
> **Research Status:** RED TEAM R&D / DEFENSIVE GAP ANALYSIS
> **Core Principle:** Multi-Tiered Transport Resilience & Temporal Evasion


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
| ✅ | 1.1 | **Transport Abstraction:** Rust Traits for hot-swappable modules. | DONE |
| ✅ | 1.2 | **Temporal Evasion:** NHPP jitter engine for non-linear beaconing. | DONE |
| ✅ | 1.3 | **C2 Bootstrap:** Gin backend + **VaporTrace Tactical UI**. | DONE |
| ✅ | 1.4 | **Malleable Profiles:** Header & JA3/S fingerprint randomization. | DONE |
| ✅ | 1.5 | **Failsafe Stack:** DNS Tunneling & ICMP/NTP Signaling. | DONE |
| ✅ | 1.6 | **Sprint 1 Finalize:** Autocomplete, README, and Integrity Commit. | DONE |
| **Sprint 2: Recon** |  |  |  |
| ✅ | 2.1 | **Artifact Harvesting:** Parsing `known_hosts` and `bash_history`. | DONE |
| ✅ | 2.2 | **Environment Context:** IMDSv2 (Cloud) & Container detection. | DONE |
| ✅ | 2.3 | **EDR/XDR Fingerprinting:** Driver enumeration & API hooks. | DONE |
| ✅ | 2.4 | **Full-Spectrum C2:** Enabling Go listeners for all 6 transport tiers. | DONE |
| ✅ | 2.5 | **Sprint 2 Finalize:** Autocomplete, README, and Integrity Commit. | DONE |
| **Sprint 3: Propagate** |  |  |  |
| ❌ | **3.1** | **RCE Framework:** Task queuing and remote `sh`/`cmd` execution. | **ACTIVE** |
| ❌ | 3.2 | **Credential Management:** Secure handling of NTLM/SSH tokens. | PLANNED |
| ❌ | 3.3 | **P2P Discovery:** mDNS/UDP gossip mesh for peer discovery. | PLANNED |
| ❌ | 3.4 | **Infection Engine:** Propagation via SMB, SSH, and WMI mocks. | PLANNED |
| ❌ | 3.5 | **Autonomous Lateral Movement:** Credential-driven "Self-Hopping". | PLANNED |
| ❌ | 3.6 | **Safety Throttle:** Rate-limiting and global "Kill-Switch". | PLANNED |
| **Sprint 4: DFIR** |  |  |  |
| ❌ | 4.1 | **LotL Persistence:** WMI Event Subs, Systemd, and GPO usage. | PLANNED |
| ❌ | 4.2 | **Phantom Memory:** Direct Syscalls & Process Injection (No-Disk). | PLANNED |
| ❌ | 4.3 | **Polymorphic Engine:** Per-hop Signature Hash Mutation. | PLANNED |
| ❌ | 4.4 | **Atomic Destruction:** Self-deletion and secure file wiping. | PLANNED |
| ❌ | 4.5 | **Anti-Forensic Scorch:** Memory-wipe & Log cleaning on Kill-Switch. | PLANNED |
| ❌ | 4.6 | **CLI Completion:** Final Shell polish and documentation audit. | PLANNED |

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

## 🔍 **Intelligence & Reconnaissance Pillars**

| Pillar | Designation | Objective | Technical Implementation |
| --- | --- | --- | --- |
| **I** | **Host Discovery** | Establish the "Target Fingerprint." | Querying `uname`, Kernel version, CPU Arch, Hostname, and UUID. |
| **II** | **User & Identity** | Map the "Human Context." | Current `whoami`, UID/GID, group memberships, and `last` login history. |
| **III** | **Artifact Harvesting** | Extract lateral movement leads. | Parsing `known_hosts`, `bash_history`, and RDP `MRU` cache. |
| **IV** | **Environment Context** | Physical vs. Cloud DNA. | IMDSv2 (AWS/Azure/GCP) probing and Container (K8s/Docker) checks. |
| **V** | **Defensive Profiling** | Identify "Predators" (EDR/AV). | Enumerating loaded drivers, syscall hooks, and security heartbeats. |
| **VI** | **Network Topology** | Map the local "Gossip" radius. | ARP table analysis, routing tables, and mDNS/LLMNR discovery. |
| **VII** | **Credential Mining** | Locate active tokens for pivoting. | Memory scraping for tokens, `.aws/credentials`, and `.kube/config`. |

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

#### **2. Multi-Tiered Network Polymorphism (Phase 1.1 - 1.5)**

The core architecture rests on a **Transport Abstraction Layer** implemented via Rust Traits, allowing the Agent to execute "Hot Swaps" of its communication protocol upon detecting network interference or egress blocking.

* **Malleable HTTPS (Phase 1.4):** Employs JA3/S fingerprint randomization and HTTP/2 header rotation. The Agent mimics legitimate browser traffic by rotating User-Agents and including randomized "Hydra-Keys" in the headers.
* **Covert Failsafes (Phase 1.5):** Includes binary-level signaling via ICMP Echo Request payloads and NTP Transmit Timestamp manipulation for low-bandwidth, high-stealth exfiltration when standard ports are closed.

#### **3. RFC-Hardened DNS Tunneling (Phase 2.2)**

Developed as the "Last Resort" transport, the DNS tunnel circumvents Deep Packet Inspection (DPI) by encapsulating telemetry within recursive DNS queries.

* **Format Integrity:** To bypass RFC 1035 constraints identified in protocol analysis, the Agent fragments Base64 payloads into 60-character labels.
* **Payload Minification:** Utilizes a single-letter JSON schema (e.g., `"a"` for `agent_id`) to ensure the total query name—including labels and root domain—remains under the strict **255-byte limit**.
* **Asynchronous Reassembly:** The Go-based Orchestrator transparently strips sub-domain delimiters and re-aligns Base64 padding before unmarshaling the telemetry.

#### **4. Intelligence Reconnaissance & Artifact Harvesting (Phase 2.1)**

The Agent implements a non-intrusive harvesting engine designed to extract lateral movement leads and environment context:

* **Environment DNA:** Implements dual-mode detection for Cloud Instance Metadata Services (IMDSv2) and containerization markers (e.g., `/.dockerenv`).
* **Credential Mining:** Specifically targets `known_hosts` and `bash_history`. To minimize the exfiltration footprint, the Agent utilizes a sliding-window buffer, capturing only the most recent interactive commands for C2 preview.
* **Memory Sanitization:** To thwart forensic RAM dumps, all sensitive telemetry structs implement the `zeroize` pattern, ensuring data is wiped from memory immediately after transport.

Excellent work, Soldier. The environment is finally battle-ready. Since we’ve successfully navigated the static linking minefield and optimized the Docker BuildKit workflow, here is the updated **LAB** section for your Whitepaper/README.

This section is designed to be scannable, technical, and aligned with your current Arch Linux environment.

---

## 🛠 LAB ENVIRONMENT

This project utilizes a high-fidelity containerized laboratory to simulate C2 communication and worm propagation.

### 1. Requirements & Tooling

To maintain the integrity of the static binaries and network isolation, the following host configuration is required:

* **Builder:** `docker-buildx` (Mandatory for modern BuildKit features).
* **Runtime:** `docker-compose` v2.x.
* **Networking:** Linux `bridge` driver with `NET_ADMIN` capabilities for raw socket manipulation (ICMP/IGMP).

### 2. Infrastructure Map

The lab is deployed on a private subnet `10.5.0.0/24`.

| Container | Role | IP Address | Capabilities |
| --- | --- | --- | --- |
| **hydra-c2-lab** | Orchestrator (C2) | `10.5.0.5` | `NET_RAW`, `NET_ADMIN` |
| **hydra-agent-alpha** | Initial Vector (Agent) | `10.5.0.10` | `NET_RAW`, `NET_ADMIN` |

### 3. Build Architecture

The `hydra-agent` is compiled using a **Multi-Stage "Forge & Ghost"** pattern to minimize footprint and eliminate external dependencies.

* **The Forge (Stage 1):** Uses `rust:alpine` with `openssl-vendored`. It compiles OpenSSL from source to ensure the binary is a fully static `musl` executable.
* **The Ghost (Stage 2):** A hardened `alpine:latest` image containing only the static binary and `ca-certificates`.

### 4. Deployment Commands

Use the provided `Makefile` to handle the environment-specific BuildKit variables:

```bash
# Build and deploy the laboratory
make up

# Verify Agent connectivity
docker exec -it hydra-agent-alpha ./hydra-agent --help

# Monitor C2 traffic
docker logs -f hydra-c2-lab

```

> **Note:** If building manually on Arch Linux, ensure you export `DOCKER_BUILDKIT=1` to avoid linker errors associated with the classic Docker builder.

---

### **IV. MITRE ATT&CK® MAPPING (SYNCHRONIZED)**

| Tactic | Technique | ID | Hydra-Worm Implementation Detail |
| --- | --- | --- | --- |
| **Reconnaissance** | Search Victim-Owned Websites | T1594 | Probing Cloud IMDSv2 (169.254.169.254) for instance identity and roles. |
| **Discovery** | System Information Discovery | T1082 | Extracting Hostname, OS Version, and Kernel details via `sysinfo`. |
| **Discovery** | File and Directory Discovery | T1083 | Targeting specific paths for exfiltration: `~/.bash_history` and `~/.ssh/known_hosts`. |
| **Discovery** | Virtualization/Sandbox Evasion | T1497 | Detection of `.dockerenv` to identify containerized constraints. |
| **Defense Evasion** | Indicator Removal | T1070 | Implementing `zeroize` patterns for in-memory telemetry and sanitizing bash history strings. |
| **Defense Evasion** | Protocol Impersonation | T1001.003 | Mimicking standard DNS traffic via manual RFC 1035 packet construction. |
| **Command & Control** | Application Layer Protocol | T1071.004 | DNS Tunneling utilizing 60-character sub-domain labels for Base64 exfiltration. |
| **Command & Control** | Traffic Signaling | T1543 | **NHPP-based** heartbeats using an exponential distribution to generate stochastic Jitter. |

---

### **V. DFIR RESPONSE TEMPLATE (NIST SP 800-61 R3)**

#### **1. Detection & Analysis (ID.AN)**

* **Network Artifacts:** Monitor for anomalous DNS query patterns. Specifically, look for high-frequency queries to a single root domain (e.g., `*.c2.hydra-worm.local`) where sub-domain labels appear to be high-entropy Base64.
* **Length Analysis:** Alerts should trigger on DNS "Null" queries or "A" records where the total QNAME length approaches the **255-byte limit**.
* **Endpoint Artifacts:** Audit for unusual process access to `.bash_history` or `.ssh/known_hosts` originating from non-interactive shells or unauthorized binaries.

#### **2. Containment, Eradication, & Recovery (PR.PT)**

* **Network Containment:** Sinkhole the authoritative nameserver for the identified C2 root domain. Implement DNS Response Policy Zones (RPZ) to block the exfiltration path.
* **Host Containment:** Isolate identified nodes. Because the Agent checks for `.dockerenv`, ensure containment does not inadvertently signal the Agent to hibernate or execute anti-forensic wipes.
* **Eradication:** Scan for the unique "Hydra-Key" in memory or HTTP headers (if Tier 2 was utilized) to identify active process injections.

---
