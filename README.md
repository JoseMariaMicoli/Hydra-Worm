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
| **Sprint 3** |  | **Propagate & Command** |  |
| ✅ | 3.1 | **RCE Framework:** Task queuing and remote `sh`/`cmd` execution. | DONE |
| ✅ | 3.2 | **Tactical Console (v4):** Cold-War TUI, µs Telemetry, & Docker Lab. | DONE |
| ✅ | 3.3 | **Credential Management:** Secure handling of NTLM/SSH tokens. | DONE |
| ✅ | 3.4 | **P2P Discovery:** mDNS/UDP gossip mesh for peer discovery. | DONE |
| ❌ | **3.5** | **Infection Engine:** Propagation via SMB, SSH, and WMI mocks. | **ACTIVE** |
| ❌ | 3.6 | **Autonomous Lateral Movement:** Credential-driven "Self-Hopping". | PLANNED |
| ❌ | 3.7 | **Safety Throttle:** Rate-limiting and global "Kill-Switch". | PLANNED |
| **Sprint 4: DFIR** |  |  |  |
| ❌ | 4.1 | **LotL Persistence:** WMI Event Subs, Systemd, and GPO usage. | PLANNED |
| ❌ | 4.2 | **Phantom Memory:** Direct Syscalls & Process Injection (No-Disk). | PLANNED |
| ❌ | 4.3 | **Polymorphic Engine:** Per-hop Signature Hash Mutation. | PLANNED |
| ❌ | 4.4 | **Atomic Destruction:** Self-deletion and secure file wiping. | PLANNED |
| ❌ | 4.5 | **Anti-Forensic Scorch:** Memory-wipe & Log cleaning on Kill-Switch. | PLANNED |
| ❌ | 4.6 | **CLI Completion:** Final Shell polish and documentation audit. | PLANNED |


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

To finalize the **Phase 3 Research Documentation**, we will break down the technical architecture of these four pillars. These details should be integrated into your **III. TECHNICAL WHITE PAPER** section to preserve the R&D integrity of the project.

---


#### **5 RCE Framework: Asynchronous Task Queuing**

The Remote Code Execution (RCE) engine is designed as a **Pull-Model Architecture** to bypass stateful firewalls.

* **Queue Mechanism**: The Orchestrator maintains a thread-safe `map[string]string` (AgentID -> Command). When an agent performs its NHPP-jittered heartbeat, the Orchestrator checks for a pending task.
* **Payload Execution**: The Rust agent receives the tasking via the JSON `task` field. It spawns a detached subprocess using `std::process::Command`, capturing `stdout` and `stderr`.
* **Response Encapsulation**: Results are prefixed with `OUT:` and re-sent in the next telemetry pulse. This ensures the C2 remains stateless and resistant to connection drops.

#### **6 Tactical Console (v4): High-Concurrency TUI**

The **VaporTrace Tactical UI** is built on the `tview` and `tcell` libraries, utilizing a multi-layered grid system.

* **µs Telemetry Processing**: The UI calculates RTT (Round Trip Time) and Jitter at the microsecond level using `time.Since(arrival).Microseconds()`. This allows for real-time detection of network throttling or interception.
* **Concurrency Model**: The Go backend uses `app.QueueUpdateDraw()`, which allows background network listeners (DNS, ICMP, NTP) to push updates to the UI thread safely without causing race conditions or flickering.
* **Docker Lab Integration**: The console is environment-aware, mapping `AgentID` to internal container metadata for localized testing within the `10.5.0.0/24` subnet.

#### **7 Credential Management: Token Isolation & Vaulting**

This module focuses on the secure exfiltration of high-value identity artifacts discovered during Phase 2.

* **Loot Ingestion Engine**: The Orchestrator identifies incoming telemetry prefixed with `LOOT:`. It parses the category (e.g., `SSH`, `AWS`, `NTLM`) and data payload.
* **De-duplication Vault**: To prevent redundant data from cluttering the database, the C2 performs a cryptographic check (comparison of the data string) before committing to the `vault` slice.
* **Zeroize Pattern**: On the Rust agent side, credentials extracted from `~/.ssh/known_hosts` or memory are stored in encrypted buffers and cleared using the `zeroize` crate immediately after the transport pulse is acknowledged.

### **8 P2P Discovery: mDNS & UDP Gossip Mesh**

The mesh capability enables "Island Hopping" in segmented networks where the primary C2 is unreachable.

* **Discovery Protocol**: Agents utilize **mDNS (Multicast DNS)** on UDP/5353 to broadcast their presence locally using the `_hydra._tcp` service type.
* **Relay Logic**: If `Agent-A` cannot reach the C2 but sees `Agent-B` (which has a Tier-1 Cloud link), `Agent-A` will encapsulate its telemetry into a POST request directed at `Agent-B:8080/relay`.
* **Broadcast Propagation**: The `broadcast` verb allows the Orchestrator to push a single command to the entire mesh. Each node that receives the command from the C2 then "gossips" it to local peers, ensuring  distribution across the network.

---

## 🛠 LAB ENVIRONMENT (PHASE 3.5 REVISION)

This project utilizes a high-fidelity containerized laboratory to simulate C2 communication, P2P mesh relay, and autonomous worm propagation.

### 1. Requirements & Tooling

* **Builder:** `docker-buildx` (Mandatory for modern BuildKit features).
* **Runtime:** `docker-compose` v2.x.
* **Networking:** Linux `bridge` driver with `NET_ADMIN` capabilities for raw socket manipulation.
* **OS Specifics:** On Arch Linux, `DOCKER_BUILDKIT=1` is required for static `musl` compilation.

### 2. Infrastructure Map (Multi-Node Mesh)

The lab is deployed on a private subnet `10.5.0.0/24`.

| Container | Role | IP Address | Status | Capabilities |
| --- | --- | --- | --- | --- |
| **hydra-c2-lab** | Orchestrator (C2) | `10.5.0.5` | ONLINE | Ports 8080, 53, 123 |
| **hydra-agent-alpha** | Patient Zero | `10.5.0.10` | **INFECTED** | mDNS, SSH Client |
| **hydra-agent-bravo** | Relay Node | `10.5.0.11` | **INFECTED** | mDNS, P2P Relay |
| **hydra-agent-gamma** | Target Node | `10.5.0.12` | **CLEAN** | SSH Server |

### 3. Operational Command Reference

#### **A. Makefile (Host Automation)**

| Command | Usage | Description |
| --- | --- | --- |
| `make up` | `make up` | Builds and deploys the 3-node laboratory and C2. |
| `make down` | `make down` | Tears down the environment and wipes virtual networks. |
| `make shell` | `make shell` | Enters the interactive VaporTrace TUI on the C2. |
| `make patch-c2` | `make patch-c2` | Recompiles and restarts the C2 without resetting agents. |
| `make logs` | `make logs` | Streams raw traffic logs (HTTP/DNS/ICMP) from the C2. |
| `make infect-gamma` | `make infect-gamma` | Manual trigger for Phase 3.5 propagation test. |

#### **B. VaporTrace TUI (Orchestrator Shell)**

| Tactical Verb | Syntax | Objective |
| --- | --- | --- |
| **exec** | `exec <ID> <CMD>` | Tasks a specific node with a shell command. |
| **broadcast** | `broadcast <CMD>` | Tasks every node in the mesh simultaneously. |
| **loot** | `loot` | Displays the vault of de-duplicated exfiltrated tokens. |
| **tasks** | `tasks` | Displays the current pending task queue. |
| **clear** | `clear` | Wipes the transmission window log. |
| **exit** | `exit` | Gracefully terminates the C2 session. |

### 4. Build Architecture

The `hydra-agent` uses a **Multi-Stage "Forge & Ghost"** pattern:

* **The Forge (Stage 1):** `rust:alpine` compiles the static `x86_64-unknown-linux-musl` binary.
* **The Ghost (Stage 2):** A hardened, minimal `alpine` image containing only the binary, mimicking a stealthy footprint.

---

### **IV. MITRE ATT&CK® MAPPING (SYNCHRONIZED - PHASE 3.5)**

| Tactic | Technique | ID | Hydra-Worm Implementation Detail |
| --- | --- | --- | --- |
| **Reconnaissance** | Search Victim-Owned Websites | T1594 | Probing Cloud IMDSv2 (169.254.169.254) for instance identity and roles. |
| **Discovery** | System Information Discovery | T1082 | Extracting Hostname, OS Version, and Kernel details via `sysinfo`. |
| **Discovery** | File and Directory Discovery | T1083 | Targeting specific paths: `~/.bash_history` and `~/.ssh/known_hosts`. |
| **Discovery** | Remote System Discovery | T1018 | Utilizing **mDNS (UDP/5353)** and ARP table analysis to map local mesh peers. |
| **Discovery** | Virtualization/Sandbox Evasion | T1497 | Detection of `/.dockerenv` to identify containerized constraints. |
| **Defense Evasion** | Indicator Removal | T1070 | Implementing `zeroize` patterns and sanitizing `bash_history` strings. |
| **Defense Evasion** | Protocol Impersonation | T1001.003 | Mimicking standard DNS/NTP/ICMP traffic via manual packet construction. |
| **Lateral Movement** | Remote Services | T1021.004 | **(ACTIVE)** Propagation via SSH utilizing harvested credentials/keys. |
| **Command & Control** | Application Layer Protocol | T1071.004 | DNS Tunneling utilizing 60-character labels for Base64 exfiltration. |
| **Command & Control** | Dynamic Resolution | T1568 | NHPP-based heartbeat intervals using stochastic jitter to bypass frequency analysis. |

---

### **V. DFIR RESPONSE TEMPLATE (NIST SP 800-61 R3)**

#### **1. Detection & Analysis (ID.AN)**

* **Network Artifacts:** Monitor for anomalous DNS query patterns (`*.c2.hydra-worm.local`). Watch for high-entropy subdomains and QNAME lengths approaching the **255-byte limit**.
* **P2P Signaling:** Detect unauthorized **mDNS (5353)** or **UDP Gossip** traffic between internal nodes, signaling lateral discovery.
* **Endpoint Artifacts:** Audit for non-interactive shells accessing `~/.ssh/known_hosts` or `/proc/net/arp`. Monitor for `nohup` execution of unknown binaries in `/tmp/`.

#### **2. Containment, Eradication, & Recovery (PR.PT)**

* **Network Containment:** Sinkhole the C2 root domain and implement **DNS Response Policy Zones (RPZ)**. Block internal port **8080** and **5353** to break the P2P relay chain.
* **Host Containment:** Isolate the "Patient Zero" (`10.5.0.10`). Be aware that killing the process may trigger the "Kill-Switch" anti-forensic routine (Phase 4.5).
* **Eradication:** Identify and remove the static `musl` binary. Use `lsof` to find hidden listeners on non-standard ports used for Tier 4/5 signaling.

---
