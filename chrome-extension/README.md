# KontentKop Chrome Extension

Psycholinguistic harm detection for any web page, powered by KontentKop's 15-metric scoring engine with IOA classification levels and p2p community voting.

## Architecture

```
┌─────────────────────────────────────────────────┐
│  Browser Page                                    │
│  ┌───────────────────────────────────────────┐  │
│  │  Content Script                            │  │
│  │  ┌─────────┐  ┌──────────┐  ┌──────────┐ │  │
│  │  │ KK 15   │→ │ Scorer   │→ │ IOA      │ │  │
│  │  │ Metrics │  │ (BC)     │  │ Classify │ │  │
│  │  └─────────┘  └──────────┘  └──────────┘ │  │
│  │       │                          │        │  │
│  │       ▼                          ▼        │  │
│  │  ┌─────────┐              ┌──────────┐   │  │
│  │  │Highlight│              │ P2P Vote │   │  │
│  │  │ + Panel │              │ Gun.js + │   │  │
│  │  │ + Badge │              │ OrbitDB  │   │  │
│  │  └─────────┘              └──────────┘   │  │
│  └───────────────────────────────────────────┘  │
└─────────────────────────────────────────────────┘
         │                          │
         ▼                          ▼
   IOA REST API              P2P Network
   (Q4 2026)              (Gun.js relays +
                           IPFS/OrbitDB)
```

## Setup

### 1. Install Dependencies

The p2p layer requires Gun.js, Helia (IPFS), and OrbitDB bundled for browser use.

```bash
cd chrome-extension
npm init -y
npm install gun helia @orbitdb/core
```

### 2. Bundle for Browser

Gun.js works directly in browsers. Helia and OrbitDB need bundling:

```bash
npx esbuild node_modules/gun/gun.js --bundle --outfile=vendor/gun.min.js --format=iife
npx esbuild vendor/helia-bundle.js --bundle --outfile=vendor/helia.min.js --format=iife
npx esbuild vendor/orbitdb-bundle.js --bundle --outfile=vendor/orbitdb.min.js --format=iife
```

Or use the pre-built CDN versions by adding them to manifest.json content_scripts.

### 3. Load Extension

1. Open `chrome://extensions`
2. Enable "Developer mode"
3. Click "Load unpacked"
4. Select the `chrome-extension/` directory

### 4. Generate Icons

Open `icons/generate_icons.html` in a browser, right-click each canvas, and save as PNG.

## P2P Stack

| Layer | Technology | Purpose |
|-------|-----------|---------|
| Real-time sync | Gun.js | Instant vote propagation between active peers |
| Persistent storage | OrbitDB | Content-addressed vote records on IPFS |
| Content addressing | IPFS (Helia) | Decentralized storage substrate for OrbitDB |
| Content ID | SHA-256(URL) | Stable identifier for associating votes with pages |

## IOA Integration

The extension maps KontentKop Body Count scores to IOA's five-level classification system:

| BC Score | IOA Level | Badge |
|----------|-----------|-------|
| >= 0.40 | CRITICAL | 🔴 OUTRAGED |
| >= 0.20 | ELEVATED | 🟠 UPSET |
| >= 0.10 | MODERATE | 🟡 CONCERNED |
| >= 0.05 | LOW | 🔵 AWARE |
| < 0.05 | CLEARED | 🟢 SAFE |

Classifications are provisional (based on local KK analysis) until the IOA REST API launches (expected Q4 2026), at which point official ITVB-certified determinations will be fetched.

Community votes via the p2p layer provide additional signal: confirmed, disputed, or mixed consensus.

## References

- [KontentKop Bake-Off](https://github.com/brettwhitty/KontentKop-Bake-Off)
- [Internet Outrage Authority](https://outrage.dataglut.org)
- [Gun.js](https://gun.eco)
- [OrbitDB](https://orbitdb.org)
- [Helia (IPFS)](https://github.com/ipfs/helia)
