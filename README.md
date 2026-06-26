<p align="center">
  <img src="resources/logo.png" alt="Water AI Logo" width="200"/>
</p>

<h1 align="center">🌊 Water AI: The AI Supermodel</h1>

<p align="center">
  <em>"You put water into a cup, it becomes the cup. You put water into a bottle, it becomes the bottle. Be water, my friend."</em> — <strong>Bruce Lee</strong>
</p>

<p align="center">
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-Apache_2.0-blue.svg" alt="License"></a>
  <a href="https://goreportcard.com/report/github.com/StellariumFoundation/Water"><img src="https://goreportcard.com/badge/github.com/StellariumFoundation/Water" alt="Go Report Card"></a>
  <a href="ROADMAP.md"><img src="https://img.shields.io/badge/Phase-1_MVP-cyan" alt="Version"></a>
  <a href="https://golang.org/dl/"><img src="https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go&logoColor=white" alt="Go Version"></a>
</p>

---

## 📖 Overview & Vision

**Water AI** is the intelligent core of the Water ecosystem — a unified **AI Supermodel** designed to aggregate the world's best specialized AI capabilities into a single, accessible platform. Conceived by **John Victor** of the **Stellarium Foundation**, it is envisioned as a practical and accessible form of **Artificial General Intelligence (AGI)** and a gift to humanity.

Water AI serves as the **brain of the ecosystem** — a universal force-multiplier for human potential and a cornerstone technology for achieving the **"Elevation to Eden."** It understands complex requests, routes them to state-of-the-art specialized AI models (finance, law, engineering, creative arts, and more), and then **acts** — drafting contracts, generating 3D designs, composing strategies, coding software, and launching campaigns.

> Water AI is not merely a tool — it is a platform that augments every human being with the combined intelligence of the world's best AI systems.

---

## 🔍 The Problem: Fragmentation of AI & Untapped Human Potential

The AI revolution is here, yet its power remains **fragmented**. Users must navigate a complex ecosystem of specialized tools — one for writing, another for coding, yet another for research, and still more for creative work. General-purpose AI models lack deep expertise in every domain, while specialized models remain siloed and inaccessible to most people.

This fragmentation creates barriers:
- **Knowledge workers** waste time switching between tools and manually synthesizing outputs
- **Professionals** in law, finance, and healthcare lack AI that truly understands their domain
- **Creatives** cannot seamlessly move from ideation to execution within a single workflow
- **Students and researchers** struggle to access the best AI capabilities for their specific needs
- **The general public** is locked out of advanced AI by complexity and cost

The result: **humanity's potential remains untapped**, limited by the tools available rather than the ambition of the user.

---

## 💡 The Solution: A Unified AI Supermodel

Water AI solves this by acting as the **Master Orchestrator**. It provides a single, fluid interface that:

1. **Understands** complex, multi-domain intent through natural language
2. **Decomposes** tasks into actionable sub-tasks
3. **Routes** each sub-task to the best-in-class specialized AI model
4. **Executes** actual digital labor — creating documents, writing code, browsing the web, manipulating data
5. **Synthesizes** results into coherent, high-quality outputs
6. **Iterates** based on user feedback, maintaining full context throughout

Water AI intelligently routes requests to specialized models across finance, law, engineering, creative arts, and more — then **acts on them**, performing real work rather than just generating text.

---

## 🚀 Key Differentiators

| Differentiator | Description |
|---|---|
| **Best-of-Kind Specialization** | Dynamically leverages a curated ecosystem of the world's leading specialized AIs. Rather than one model trying to do everything, Water AI routes each task to the model that does it best — creating a dynamic ecosystem that evolves as new models emerge. |
| **True Action & Labor Performing** | Goes beyond text generation — creates, builds, and executes across digital tasks including document creation, code generation, web interaction, data manipulation, and creative media generation. |
| **Open, Accessible & Potentially Client-Run** | Open-source core components with the ability to run on client devices via WebAssembly for data sovereignty and privacy. Designed to be accessible to everyone, not just enterprises. |
| **Intelligent Orchestration** | Sophisticated Go-based AI core that plans multi-step workflows, selects the right tools and models, seeks clarification when needed, and maintains context across complex interactions. |
| **Formless & Persistent** | Runs as a background daemon on your OS, accessible via a native desktop GUI, web interface, or remote bridges — always ready, always adapting. |

---

## 🎯 Target Users

- **Professionals** — Legal, Finance, Healthcare, Engineering, Business strategists who need domain-expert AI
- **Creatives** — Designers, Writers, Artists, Developers who need seamless ideation-to-execution workflows
- **Researchers & Academics** — Deep research, fact-checking, data analysis, and knowledge synthesis
- **Students & Lifelong Learners** — Learning augmentation, tutoring, and knowledge exploration
- **General Knowledge Workers & Individuals** — Everyday productivity, task automation, and personal AI assistance

---

## ✨ Product Features

### 🖥 Intuitive Multi-Modal Interface
- Chat interface supporting **text, file uploads, and voice input**
- **Downloadable native desktop client** (Windows, macOS, Linux) built with [Fyne](https://fyne.io/)
- **Web access** via embedded frontend
- Multi-format output with user control and iteration
- Browser, code, and terminal panels for real-time visualization of AI actions

### 🧠 Intelligent Prompt Processing & Orchestration Engine
- Intent understanding and task decomposition
- Multi-model routing to specialized AIs
- Context management with token-aware windowing and conversation summarization
- Response synthesis across multiple model outputs
- Sequential thinking and planning modules
- Quality assurance via reviewer agent
- Slash commands for quick actions (`/help`, `/compact`)

### 🤖 Specialized AI Model Ecosystem
- Curated models from **Hugging Face, open-source, and commercial sources**
- Multi-provider LLM support: **OpenAI, Anthropic (Claude), Google Gemini**
- Domain-specific routing for finance, law, engineering, creative arts, and more
- Chain-of-thought reasoning support (o1/o3 models, Anthropic thinking tokens)
- Configurable model parameters (temperature, retries, token limits)

### ⚡ Action & Labor Performing Engine
- **Document Creation** — Drafting contracts, reports, presentations
- **Code Generation & Execution** — Write, run, and debug software in sandboxed environments (Docker, E2B, local)
- **Web Interaction** — Browser automation via Playwright for research, data gathering, and web tasks
- **Data Manipulation** — Processing, analysis, and visualization
- **Creative Generation** — Image generation, audio transcription, media processing
- **Client-side execution** (WebAssembly/Python) and **cloud-side execution** paths

### 🔗 Integration Framework
- **Web search**: Tavily, Jina, SerpAPI, DuckDuckGo
- **Web scraping**: Firecrawl integration
- **Third-party APIs**: Vercel, NeonDB, cloud storage
- **Cloud platforms**: Google Cloud Platform, Azure
- **Planned**: Email, social media, calendar, CRM, project management, and web service connectors

### 📤 Output Management & Refinement
- Multi-format output rendering (text, code, images, files)
- Iterative refinement through conversation
- File management with workspace-per-session isolation
- Export and download capabilities

---

## 🏗 Technical Architecture

```
Water AI Architecture
┌─────────────────────────────────────────────────────────┐
│                    CLIENT LAYER                          │
│  ┌──────────┐  ┌──────────────┐  ┌───────────────────┐  │
│  │ Fyne GUI │  │ Web Frontend │  │ WebAssembly (WASM) │  │
│  └────┬─────┘  └──────┬───────┘  └────────┬──────────┘  │
│       └───────────┬────┘                   │             │
│              WebSocket / HTTP              │             │
└──────────────────┬─────────────────────────┘             │
                   │                                       │
┌──────────────────▼───────────────────────────────────────┐
│                   SERVER LAYER                            │
│  ┌─────────────────────────────────────────────────────┐ │
│  │           Gateway (Gin HTTP/WS Server)              │ │
│  │  • Session Management  • File Upload  • Health API  │ │
│  └────────────────────┬────────────────────────────────┘ │
│                       │                                  │
│  ┌────────────────────▼────────────────────────────────┐ │
│  │          Orchestration Engine (Agents)               │ │
│  │  • Prompt Builder  • Context Manager  • Reviewer    │ │
│  │  • Task Decomposition  • Tool Selection             │ │
│  └────────────────────┬────────────────────────────────┘ │
│                       │                                  │
│  ┌────────────────────▼────────────────────────────────┐ │
│  │              Tool & Action Layer                     │ │
│  │  • Browser  • Terminal  • File Editor  • Search     │ │
│  │  • Media    • Code Exec • Web Scraping              │ │
│  └────────────────────┬────────────────────────────────┘ │
│                       │                                  │
│  ┌────────────────────▼────────────────────────────────┐ │
│  │           LLM Provider Layer                        │ │
│  │  • OpenAI  • Anthropic  • Google Gemini             │ │
│  └─────────────────────────────────────────────────────┘ │
│                                                          │
│  ┌─────────────────────────────────────────────────────┐ │
│  │  Infrastructure: SQLite/GORM • Sandbox (Docker/E2B) │ │
│  │  Config • Logging • Migrations • Process Manager    │ │
│  └─────────────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────────┘
```

**Client-side:** Native Fyne desktop GUI + WebAssembly runtime for local execution  
**Server-side:** Go-based orchestration engine with multi-provider LLM routing  
**Privacy by Design:** Minimal data retention with client-side execution options. Envisioned to run high-performance actions locally on client devices to ensure data sovereignty.

---

## 📁 Project Structure

```
Water/
├── cmd/water/              # Main entry point (daemon + Fyne GUI launcher)
├── server/                 # HTTP/WebSocket server (Gin-based)
├── agents/                 # AI agent abstractions (Base, Reviewer, FunctionCall)
├── browser/                # Headless browser automation (Playwright)
├── core/                   # Core utilities: logging, config, event system, storage
├── db/                     # Database layer (SQLite via GORM)
├── llm/                    # LLM clients (Anthropic, Gemini, OpenAI)
│   └── context_manager/    # Token counting and context window management
├── migrations/             # Database migrations (Goose)
├── process/                # Process/session management & gateway
├── prompts/                # System prompt builder with domain-specific rules
├── sandbox/                # Sandboxed execution (Docker, E2B, Local)
├── tools/                  # Tool implementations (browser, terminal, search, media, etc.)
├── ui/                     # Fyne desktop GUI (chat, panels, settings, theme)
├── utils/                  # Shared utilities (file manager, terminal manager)
├── resources/              # Logo and static assets
├── plans/                  # Architecture and planning documents
├── .github/                # CI/CD workflows (GitHub Actions)
├── Makefile                # Unified build system
└── ROADMAP.md              # Detailed feature roadmap
```

---

## 💰 Business Model

| Revenue Stream | Description |
|---|---|
| **Community Donations** | Voluntary donations from individuals and organizations who benefit from Water AI |
| **Enterprise Licensing** | Volume licensing and tailored solutions for businesses requiring dedicated support and SLAs |
| **Usage-Based Pricing** | Fair pricing based on inference and cloud compute consumption for heavy users |

---

## 🏛 The Visionary

| | |
|---|---|
| **Visionary** | **John Victor** |
| **Organization** | Stellarium Foundation |
| **Mission** | Leverage technology for global prosperity and human advancement |
| **Goal** | The Elevation to Eden |

John Victor, founder of the **Stellarium Foundation**, conceived Water AI as the technological cornerstone of a broader mission to uplift humanity. The Foundation's work spans multiple initiatives, with Water AI serving as the intelligent core that ties them together.

---

## 💵 Funding

**Funding Ask:** **$600,000** for core platform development, foundational model integration, and key engineering talent.

Funds will be allocated toward:
- **Core Orchestration Engine** — Advanced task decomposition, multi-model routing, and planning capabilities
- **Foundational Model Integrations** — Connecting 500+ specialized AI models across every domain
- **Key Engineering Talent** — Hiring world-class Go, AI/ML, and systems engineers
- **Infrastructure** — Cloud compute for development, testing, and initial deployment

---

## 🛠 Getting Started

### Prerequisites

- [Go 1.24+](https://golang.org/dl/)
- GCC / Clang (CGO is required for the Fyne GUI)
  - **Linux:** `libgl1-mesa-dev`, `libxcursor-dev`, `libxrandr-dev`, `libxinerama-dev`, `libxi-dev`, `libxxf86vm-dev`
  - **macOS:** Xcode Command Line Tools
  - **Windows:** MinGW-w64

### Quick Start (Development)

```bash
git clone https://github.com/StellariumFoundation/Water.git
cd Water

# Build the Go binary (no frontend/Node.js required)
make build-dev

# Run the server in headless mode
./bin/water-ai server
```

The server starts on `http://localhost:7777`.

### Full Build (with GUI)

```bash
# Build the complete application with Fyne GUI
make build

# Run — launches gateway + native desktop GUI
./bin/Water

# Or run headless server only
./bin/Water server

# Check version
./bin/Water --version
```

### Running Tests

```bash
make test              # Run all unit tests
make test-race         # Run tests with race detection
make test-coverage     # Generate HTML coverage report
```

### Cross-Platform Release

```bash
make release           # Build optimized binaries for all platforms
```

Release binaries are output to the `dist/` directory (Linux, macOS, Windows — amd64 + arm64).

---

## 🗺 Roadmap

See [**ROADMAP.md**](ROADMAP.md) for the detailed, feature-level roadmap with implementation status.

### Phase 1: The Drop (MVP) — *In Progress*
- Core platform, multi-LLM orchestration, desktop GUI, tool framework

### Phase 2: The Stream (Expansion)
- Public API, 500+ specialized model integrations, MCP marketplace

### Phase 3: The Ocean (Global Scale)
- Community-driven marketplace, autonomous "Eden" workflows, full AGI capabilities

---

## 💰 Business Model

| Revenue Stream | Description |
|---|---|
| **Community Support** | Voluntary donations from individuals and organizations who believe in the mission |
| **Enterprise Solutions** | Volume licensing and tailored services for businesses requiring dedicated support |
| **Usage-Based Options** | Fair pricing based on inference and cloud compute consumption |

---

## 💵 Funding

**Funding Ask:** **$600,000** for core platform development, foundational model integration, and key engineering talent.

Funds will be allocated toward:
- **Core orchestration engine development** — Building the intelligent routing and task decomposition system
- **Foundational specialized model integrations** — Connecting to the best AI models across every domain
- **Key engineering talent acquisition** — Hiring world-class Go, AI/ML, and systems engineers
- **Infrastructure and cloud compute** — Development, testing, and production environments

---

## 🔓 Open Source Strategy

Water AI's core components are **open-source** under the Apache 2.0 License. This strategy ensures:

- **Transparency** — Full visibility into how the AI operates
- **Trust** — Community-auditable codebase
- **Community Contribution** — Developers worldwide can contribute and extend
- **Rapid Adoption** — No barriers to entry for individuals and organizations

---

## 🏛 The Visionary & Foundation

| | |
|---|---|
| **Visionary** | **John Victor** — Founder of the Stellarium Foundation, architect of the Water AI vision |
| **Organization** | **Stellarium Foundation** |
| **Mission** | Leverage technology for global prosperity and human advancement |
| **Goal** | The **Elevation to Eden** — using AI to unlock humanity's full potential |

John Victor conceived Water AI not as a commercial product, but as a **gift to humanity** — a platform that democratizes access to the world's most powerful AI capabilities, ensuring that every person on Earth can benefit from the AI revolution.

---

## 🤝 Contributing

Water AI is an open-source project fostered by the **Stellarium Foundation**. We welcome developers who share the vision of human augmentation.

1. Fork the repo
2. Create your feature branch (`git checkout -b feature/AmazingAction`)
3. Commit your changes
4. Push to the branch
5. Open a Pull Request

---

## 📄 License

This project is licensed under the **Apache 2.0 License** — see the [LICENSE](LICENSE) file for details.

---

<p align="center"><em>"Be water. Flow into the future."</em></p>
