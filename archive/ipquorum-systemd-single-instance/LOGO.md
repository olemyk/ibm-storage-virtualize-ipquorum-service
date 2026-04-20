# IBM Storage Virtualize IP Quorum Service - Visual Identity

This document contains ASCII art logos and banners for the IP Quorum service project.

## Main Logo - Large

```
╔═══════════════════════════════════════════════════════════════════════════╗
║                                                                           ║
║   ██╗██████╗      ██████╗ ██╗   ██╗ ██████╗ ██████╗ ██╗   ██╗███╗   ███╗  ║
║   ██║██╔══██╗    ██╔═══██╗██║   ██║██╔═══██╗██╔══██╗██║   ██║████╗ ████║  ║
║   ██║██████╔╝    ██║   ██║██║   ██║██║   ██║██████╔╝██║   ██║██╔████╔██║  ║
║   ██║██╔═══╝     ██║▄▄ ██║██║   ██║██║   ██║██╔══██╗██║   ██║██║╚██╔╝██║  ║
║   ██║██║         ╚██████╔╝╚██████╔╝╚██████╔╝██║  ██║╚██████╔╝██║ ╚═╝ ██║  ║
║   ╚═╝╚═╝          ╚══▀▀═╝  ╚═════╝  ╚═════╝ ╚═╝  ╚═╝ ╚═════╝ ╚═╝     ╚═╝  ║
║                                                                           ║
║              IBM Storage Virtualize - IP Quroum Service                   ║
║                                                                           ║
╚═══════════════════════════════════════════════════════════════════════════╝
```

## Compact Logo

```
┌─────────────────────────────────────────────────────────┐
│  _____ ____     ___                                     │
│ |_   _|  _ \   / _ \ _   _  ___  _ __ _   _ _ __ ___    │
│   | | | |_) | | | | | | | |/ _ \| '__| | | | '_ ` _ \   │
│   | | |  __/  | |_| | |_| | (_) | |  | |_| | | | | | |  │
│   |_| |_|      \__\_\\__,_|\___/|_|   \__,_|_| |_| |_|  │
│                                                         │
│        IBM Storage Virtualize HA Service                │
└─────────────────────────────────────────────────────────┘
```

## Icon/Badge Style

```
    ╔═══════════════════╗
    ║                   ║
    ║    ┌─────────┐    ║
    ║    │ IP-Q    │    ║
    ║    │  ╱╲╱╲   │    ║
    ║    │ ╱  ╲  ╲ │    ║
    ║    │╱    ╲  ╲│    ║
    ║    └─────────┘    ║
    ║   Quorum Service  ║
    ║                   ║
    ╚═══════════════════╝
```

## Conceptual Diagram Logo

```
┌────────────────────────────────────────────────────────────────┐
│                                                                │
│        Storage Cluster A  ←→  IP Quorum  ←→  Storage Cluster B │
│              ╱ ╲                  │                 ╱ ╲        │
│             ╱   ╲                 │                ╱   ╲       │
│            ╱     ╲                │               ╱     ╲      │
│           ╱  HA   ╲          Tie-Breaker         ╱  HA   ╲     │
│          ╱_________╲              │             ╱_________╲    │
│                                   │                            │
│                          [Decision Maker]                      │
│                                                                │
│              IBM Storage Virtualize IP Quorum Service          │
│                    High Availability Solution                  │
│                                                                │
└────────────────────────────────────────────────────────────────┘
```

## Network Topology Logo

```
            ┌─────────────────────┐
            │   IP Quorum Host    │
            │  ┌───────────────┐  │
            │  │  Java Service │  │
            │  │  Port: 1260   │  │
            │  └───────┬───────┘  │
            └──────────┼──────────┘
                       │
                ┌──────────────
                │              │        
                ▼              ▼        
    ┌───────────────┐  ┌───────────────┐
    │  Cluster1     │  │  Cluster2     │
    │  ┌─────────┐  │  │  ┌─────────┐  │
    │  │         │  │  │  │         │  │
    │  │ Client  │  │  │  │ Client  │  │
    │  └─────────┘  │  │  └─────────┘  │
    └───────────────┘  └───────────────┘
         Storage            Storage     
       Virtualize         Virtualize    
```

## Minimalist Logo

```
    ╭──────────╮
    │ IP-Q     │
    │ ◆─◆─◆    │  IBM Storage Virtualize
    │  \│/     │  Quorum Service
    │   ◆      │
    ╰──────────╯
```

## Banner Style - Full Width

```
╔══════════════════════════════════════════════════════════════════════════════╗
║                                                                              ║
║     ██╗██████╗       ██████╗ ██╗   ██╗ ██████╗ ██████╗ ██╗   ██╗███╗   ███╗  ║
║     ██║██╔══██╗     ██╔═══██╗██║   ██║██╔═══██╗██╔══██╗██║   ██║████╗ ████║  ║
║     ██║██████╔╝     ██║   ██║██║   ██║██║   ██║██████╔╝██║   ██║██╔████╔██║  ║
║     ██║██╔═══╝      ██║▄▄ ██║██║   ██║██║   ██║██╔══██╗██║   ██║██║╚██╔╝██║  ║
║     ██║██║          ╚██████╔╝╚██████╔╝╚██████╔╝██║  ██║╚██████╔╝██║ ╚═╝ ██║  ║
║     ╚═╝╚═╝           ╚══▀▀═╝  ╚═════╝  ╚═════╝ ╚═╝  ╚═╝ ╚═════╝ ╚═╝     ╚═╝  ║
║                                                                              ║
║  ┌────────────────────────────────────────────────────────────────────────┐  ║
║  │  IBM Storage Virtualize High Availability Quorum Service               │  ║
║  │  Automated Download • Systemd Integration                              │  ║
║  └────────────────────────────────────────────────────────────────────────┘  ║
║                                                                              ║
╚══════════════════════════════════════════════════════════════════════════════╝
```

## Service Status Display

```
┌─────────────────────────────────────────────────────────────┐
│  IP Quorum Service Status                                   │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│   ●  Service: ACTIVE                                        │
│   ●  Port: 1260/TCP                                         │
│   ●  Cluster: Connected                                     │
│   ●  Download: Enabled                                      │
│                                                              │
│        ┌───┐                                                │
│        │ Q │  ←→  Storage Virtualize Cluster               │
│        └───┘                                                │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

## Workflow Icon

```
    ┌─────────────────────────────────────┐
    │  Download → Install → Configure     │
    │     ↓          ↓          ↓         │
    │  ┌─────┐   ┌─────┐   ┌─────┐       │
    │  │ JAR │ → │ SVC │ → │ RUN │       │
    │  └─────┘   └─────┘   └─────┘       │
    │                                     │
    │    IP Quorum Automated Setup       │
    └─────────────────────────────────────┘
```

## Split-Brain Prevention Concept

```
    ┌──────────────────────────────────────────────────────┐
    │                                                       │
    │   Site A                    Site B                   │
    │   ┌────┐                    ┌────┐                   │
    │   │ ██ │ ←─────────────────→│ ██ │                   │
    │   └────┘    Link Failure    └────┘                   │
    │      ↓                         ↓                      │
    │      │                         │                      │
    │      └────────→ ┌────┐ ←───────┘                     │
    │                 │ Q  │  IP Quorum                    │
    │                 └────┘  Decides!                     │
    │                                                       │
    │        Prevents Split-Brain Scenarios                │
    │                                                       │
    └──────────────────────────────────────────────────────┘
```

## Technology Stack Badge

```
╔═══════════════════════════════════════╗
║  IP Quorum Service - Tech Stack       ║
╠═══════════════════════════════════════╣
║                                       ║
║  ┌─────────────────────────────────┐  ║
║  │ • Go Binary (Fast & Efficient)  │  ║
║  │ • Python Script (Feature-Rich)  │  ║
║  │ • Bash Script (Portable)        │  ║
║  │ • Systemd Integration           │  ║
║  │ • REST API Client               │  ║
║  │ • Automatic Updates             │  ║
║  └─────────────────────────────────┘  ║
║                                       ║
╚═══════════════════════════════════════╝
```

## Simple Header

```
═══════════════════════════════════════════════════════════
  IP-QUORUM  │  IBM Storage Virtualize HA Service
═══════════════════════════════════════════════════════════
```

## Compact Service Badge

```
┏━━━━━━━━━━━━━━━━━━━━━━┓
┃  ⚡ IP Quorum         ┃
┃  ━━━━━━━━━━━━━━━━━   ┃
┃  ✓ Auto Download     ┃
┃  ✓ Systemd Ready     ┃
┃  ✓ HA Enabled        ┃
┗━━━━━━━━━━━━━━━━━━━━━━┛
```

## Usage Examples

### In README.md Header
Use the "Banner Style - Full Width" or "Main Logo - Large"

### In Terminal Output
Use the "Compact Logo" or "Simple Header"

### In Documentation
Use the "Icon/Badge Style" or "Minimalist Logo"

### In Status Messages
Use the "Service Status Display"

### In Architecture Docs
Use the "Network Topology Logo" or "Conceptual Diagram Logo"

---

**Color Suggestions** (for terminals that support color):
- Primary: Blue (#0F62FE) - IBM Blue
- Secondary: Cyan (#00B4D8) - Network/Connectivity
- Accent: Green (#24A148) - Success/Active
- Warning: Yellow (#F1C21B) - Attention
- Error: Red (#DA1E28) - Critical

**Font Recommendations** (for graphical versions):
- Primary: IBM Plex Sans
- Monospace: IBM Plex Mono
- Alternative: Roboto, Open Sans
