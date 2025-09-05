# Cyanotype

**Cyanotype** is an open-source toolchain for managing **Bills of Materials (BoMs) as Code**.
It brings modern software practices—version control, reproducible builds, and declarative configuration—to the world of physical product development.

Cyanotype is part of the [Tychonis](https://tychonis.com) ecosystem, which is building an alternative to traditional PLM systems with openness, modularity, and developer-first ergonomics.

---

## ✨ Why Cyanotype?

- **BoM as Code**
  Author and manage engineering, manufacturing, and service BoMs using a declarative DSL inspired by Terraform/HCL.

- **Unified BoM**
  Keep EBOM, MBOM, and Service BoM consistent in one system—no more disconnected spreadsheets or fragile PLM exports.

- **Versioned & Reproducible**
  Store BoMs in Git, diff across commits, rebuild at any revision, and track supersession/lineage of parts.

- **Extensible**
  Designed to integrate with CAD, supply chain data, manufacturing processes, and even external Git repos.

- **Open Source**
  Unlike closed PLM systems, Cyanotype is transparent, hackable, and community-driven.

---

## ⚠️ Disclaimer

Cyanotype is in a very early stage of development.
The DSL, data models, and APIs are subject to rapid iteration and change.
Expect breaking changes until we approach a stable release. We encourage experimentation, feedback, and contributions—but do not rely on Cyanotype yet for production-critical workloads.

## 🚀 Getting Started

### Prerequisites
- Go 1.22+
- Git
- (Optional) PostgreSQL for persistence

### Build
```bash
git clone https://github.com/tychonis/cyanotype.git
cd cyanotype
go build .
```

## Example: Define a BoM

Create a file skateboard.bpo:

```
item "deck" {
    part_number = "D-1001"
}

item "wheel" {
    part_number = "W-2001"
}

item "assembly" {
    from = [
        {
            name = "deck"
            ref = deck
            qty = 1
        },
        {
            name = "wheel"
            ref = deck
            qty = 4
        },
    ]
}
```

Run bom:
```
./cyanotype bom skateboard.bpo assembly
```

Build intermediate state:
```
./cyanotype build skateboard.bpo assembly
```

We also created a few examples:
- [Chess](https://github.com/tychonis/cyanotype-chess)
- [Factorio](https://github.com/tychonis/cyanotype-factorio)
- [Pick and Place Machine](https://github.com/tychonis/cyanotype-pnp)

The rendering can be found in [Bom hub](https://bomhub.tychonis.com).
