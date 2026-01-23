
# HYDRA-WORM:

## I. PROJECT README & OPERATIONAL ROADMAP

### **Project Overview**

**Hydra-Worm** is a next-generation, research-oriented breach simulation framework. It utilizes a **Rust-based Agent** for low-level stealth and memory safety, and a **Go-based Orchestrator** for high-concurrency C2 operations. The framework simulates "Living off the Land" (LotL) propagation, polymorphic C2 evasion, and cross-platform lateral movement.

### **Legal Disclaimer & Rules of Engagement (RoE)**

> **CRITICAL LEGAL NOTICE:** This software is developed exclusively for **authorized Red Team Research, Development (R&D), and Defensive Gap Analysis**.
> 1. **Authorization:** Execution is strictly prohibited on any system without explicit, written "Stay Out of Jail" documentation.
> 2. **Environment Isolation:** The simulation must be restricted to logically or physically segmented lab environments. Propagation modules must be configured with a "Maximum Hop" count and "Subnet Mask" restriction to prevent accidental spillover.
> 3. **Resource Management:** Propagation and heartbeat intervals must be throttled to prevent Denial of Service (DoS) on network infrastructure or SIEM/Logging pipelines.
> 4. **Cleanup Guarantee:** Every iteration must include a pre-verified, automated "Nuclear" cleanup routine that removes all binaries, persistence keys, and log entries.
> 
> 

### **Full Project Roadmap (Sprints & Sub-Phases)**

| Phase | Sub-Phase | Focus / Technical Deliverable |
| --- | --- | --- |
| **Sprint 1: Stealth** 
| [x] | 1.1 | **Transport Abstraction:** Rust Traits for hot-swappable communication modules. |
| [ ] | 1.2 | **Temporal Evasion:** NHPP (Non-Homogeneous Poisson Process) jitter engine. |
| [ ] | 1.3 | **C2 Bootstrap:** Go-based Gin/Echo backend with TLS 1.3 and mutual auth. |
| [ ] | 1.4 | **Malleable Profiles:** Dynamic HTTP/2 header and JA3/S fingerprint randomization. |
| **Sprint 2: Recon** 
| [ ] | 2.1 | **Artifact Harvesting:** Parsing `known_hosts`, RDP `MRU`, and `bash_history`. |
| [ ] | 2.2 | **Environment Context:** IMDSv2 (AWS/Azure/GCP) and Container (K8s/Docker) detection. |
| [ ] | 2.3 | **EDR/XDR Fingerprinting:** Enumerating drivers and hooked APIs for evasion logic. |
| [ ] | 2.4 | **Structured Telemetry:** Protobuf-encoded reporting for minimal network signature. |
| **Sprint 3: Propagation** 
| [ ] | 3.1 | **Credential Management:** Secure handling and reuse of captured NTLM/Kerberos/SSH tokens. |
| [ ] | 3.2 | **P2P Discovery:** mDNS/UDP/LLMNR gossip mesh for internal peer discovery. |
| [ ] | 3.3 | **Infection Engine:** Multithreaded propagation via SMB, SSH, and WMI mocks. |
| [ ] | 3.4 | **Safety Throttle:** Propagation rate-limiting and global "Kill-Switch" broadcast. |
| **Sprint 4: DFIR** 
| [ ] | 4.1 | **LotL Persistence:** Implementation via WMI Event Subs, Systemd timers, and GPO. |
| [ ] | 4.2 | **Syscall Evasion:** Refactoring core logic for Direct/Indirect Syscalls (bypassing `ntdll`). |
| [ ] | 4.3 | **Atomic Destruction:** Self-deletion logic including secure file wiping ( method). |
| [ ] | 4.4 | **CLI Completion:** Integration of help-autocomplete, shell commands, and README update. |

---

## II. MITRE ATT&CK® MAPPING (DETAILED)

| Tactic | Technique | ID | Hydra-Worm Implementation Detail |
| --- | --- | --- | --- |
| **Reconnaissance** | Active Scanning | T1595 | Passive ARP/mDNS sniffing to identify peers without active pings. |
| **Execution** | Shared Modules | T1129 | Polymorphic transport logic loaded via reflective DLL/SO injection. |
| **Persistence** | Scheduled Task/Job | T1053 | Use of `at`, `schtasks`, and `systemd` for recurring execution. |
| **Privilege Esc.** | Valid Accounts | T1078 | Reuse of harvested tokens/keys for vertical and lateral movement. |
| **Defense Evasion** | Direct System Calls | T1562.001 | Bypassing EDR shims by invoking kernel syscalls directly in Rust. |
| **Defense Evasion** | Indicator Removal | T1070 | Zeroizing C2 metadata in RAM via the `zeroize` crate and `Drop` trait. |
| **Discovery** | System Network Config | T1016 | Mapping internal topology via `ip neighbor` and IMDS metadata. |
| **Lateral Movement** | Remote Services | T1021 | Propagation using legitimate SSH/SMB protocols with valid creds. |
| **C2** | Traffic Signaling | T1543 | **NHPP-based** heartbeats to mimic stochastic user behavior. |
| **C2** | Protocol Impersonation | T1001.003 | Encapsulating C2 traffic within legitimate AWS/GitHub API calls. |
| **C2** | Multi-hop Proxy | T1090 | Routing traffic from restricted subnets through internet-facing peers. |

---

## III. TECHNICAL WHITE PAPER: ADVANCED PERSISTENT WORM ARCHITECTURE

### **1. Executive Summary**

Hydra-Worm is designed to address the "Observability Gap" in modern SOCs. By decoupling the agent's core logic from its delivery mechanism, we achieve a **Polymorphic Lifecycle**. This white paper details the implementation of temporal evasion, transport mutation, and kernel-level bypasses.

### **2. Software Architecture & Tech Stack**

* **The Agent (Rust):** Chosen for zero-runtime overhead and deterministic memory management. Rust allows the implementation of "Unsafe" blocks to perform direct register manipulation and syscalls without the predictability of C-runtime (`msvcrt.dll` or `libc`) signatures.
* **The Orchestrator (Go):** Utilizes an asynchronous, event-driven model. The C2 backend uses Go's `net/http2` to support long-lived, multiplexed streams, mimicking legitimate modern web application behavior.

To include the **Non-Homogeneous Poisson Process (NHPP)** equations and descriptions in your `README.md`, you should use a combination of **LaTeX** (supported by GitHub and most Markdown renderers) and structured lists.

Since we are following your workflow, you can use the implementation below. After you verify this looks correct, we can move to the next step of updating your actual file.

---

Since the previous LaTeX formatting might have failed due to how Sublime Text or your specific Markdown plugin handles delimiters, let's use a "bulletproof" version. GitHub and most modern editors prefer the `$$` block for centered math.

Here is the implementation for your `README.md`. I've also included the **integrated intensity** formula which defines .

---

### Mathematical Foundation: NHPP

The probability of $n$ beacons in the interval $(t, t+\tau]$ is given by:

$$
P[N(t+\tau)-N(t)=n] = \frac{[\Lambda(t,\tau)]^n}{n!} e^{-\Lambda(t,\tau)}
$$

Where the intensity $\Lambda(t,\tau)$ is defined as:

$$
\Lambda(t,\tau) = \int_{t}^{t+\tau} \lambda(s)ds
$$

**Intensity Factors ($\lambda$):**
* **System Noise:** $\lambda$ increases during high disk I/O to blend with background traffic.
* **Time of Day:** $\lambda$ follows a sinusoidal curve to mimic office hours.

---

### **4. Network Polymorphism: The Transport Abstraction Layer**

The agent manages a registry of **Transport Providers**. A internal `Decision Engine` monitors egress health. If a specific transport (e.g., HTTPS to a CloudFront domain) triggers an EDR alert or a TCP Reset, the agent executes a "Hot Swap" of the interface.

* **API-Based (Option 2):** Data is exfiltrated via "Dead-Drop" Resolvers. The agent posts base64-encoded, encrypted blobs into legitimate cloud metadata services (e.g., GitHub Gists, Discord Webhooks).
* **P2P Gossip (Option 3):** Uses a **Kademlia Distributed Hash Table (DHT)** over UDP. If the agent loses internet egress, it seeks local peers via mDNS/LLMNR and "shuttles" its data to a peer that still has an active egress path.

### **5. Endpoint Stealth: Direct Syscalls & Memory Zeroization**

* **Syscall Evasion:** Most EDRs hook `ntdll.dll` to monitor API calls like `NtCreateThreadEx`. Hydra-Worm avoids these hooks by dynamically resolving syscall numbers from the disk-based version of `ntdll.dll` and executing them directly via assembly:
```asm
mov r10, rcx
mov eax, [syscall_number]
syscall

```


* **Memory Sanitization:** To thwart post-compromise forensic memory dumps, the agent implements the `zeroize` trait. Upon any state transition or destruction, sensitive structs (keys, C2 URLs, harvested creds) are overwritten using `volatile` operations to ensure the compiler does not optimize away the sanitization.

---

## IV. DFIR RESPONSE TEMPLATE (NIST SP 800-61 R3)

### **1. Detection (Preparation & Identification)**

* **Network:** Identify TLS handshakes with anomalous JA3/S fingerprints. Look for high-frequency UDP/5353 (mDNS) traffic in environments where it is traditionally disabled.
* **Endpoint:** Monitor for the "Fork and Run" pattern or the loading of signed binaries into unusual processes (`msiexec.exe` spawning a shell).

### **2. Analysis (RS.AN)**

* **Memory:** Extract the process environment block (PEB). Look for evidence of **Reflective Loading** and anomalous thread start addresses not backed by a disk-based module.
* **Traffic:** Use **Inter-Arrival Time (IAT)** analysis. While Poisson-distributed, the framework may still show a specific "lambda" fingerprint over a long-term (24h+) capture.

### **3. Containment & Eradication**

* **Containment:** Implement "Micro-segmentation". The primary goal is to break the mDNS/UDP peer discovery to stop the P2P mesh from forming.
* **Eradication:** Verify the removal of Scheduled Tasks and WMI Event Consumers. Note that the framework may use "Phantom Tasks" (tasks that appear deleted but remain in the registry hive).

---