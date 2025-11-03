# TerraScope

A web app that turns your **Terraform state into an interactive constellation**.  
Resources become planets, modules become galaxies, and dependencies form orbital links.

---

## Goal

Build a **stateless web microservice** that parses `terraform.tfstate` (or a remote backend) and renders an interactive 2D or 3D map of your infrastructure.  
The goal is to learn, explain, and audit infra relationships through an intuitive, explorable UI.

> These are example ideas and guidelines.  
> Your team can implement the concept however you want.

---

## Technical Requirements

- **Architecture**: stateless web microservice  
- **Inputs**: Terraform state from:
  - Local file upload `terraform.tfstate`
  - Remote backends (S3, GCS, AzureRM) via read-only credentials
  - Optional: `terraform graph -json` or plan JSON for richer edges
- **Stack (examples)**:
  - Backend: `Go / FastAPI / Node.js / Rust`
  - Frontend: `React / Svelte / Three.js / WebGL / D3`
  - Storage: optional object store for snapshots
- **CI/CD**: GitHub Actions (or similar) and **GitOps-ready** manifests (Argo CD or Flux)
- **Security**: handle secrets safely, support redaction rules (tags, attributes)

---

## Concept Ideas

> Feel free to change names, visuals, or rules.

### Visual Mappings
| Terraform Concept | TerraScope Metaphor | Example Visualization |
|-------------------|---------------------|-----------------------|
| Provider          | Star type           | Color or spectrum per provider |
| Module            | Galaxy              | Grouped cluster with faint halo |
| Resource          | Planet              | Size by importance or cost hint |
| Data source       | Moon                | Smaller orbiting body |
| Dependency edge   | Orbit/Link          | Line with arrow and weight |
| Drift or taint    | Storm cloud         | Particle effect or outline |

### Views
- **Constellation view**: force-directed graph of resources and modules  
- **Orbit view**: module-centric “solar system” where resources orbit modules  
- **Timeline**: play changes across commits or releases  
- **Policy lens**: highlight resources violating tagging or compliance rules

### Interactions
- Click to open a resource panel: type, provider, attributes (sanitized), lifecycle, dependencies  
- Filter by provider, module, tag, environment  
- Search bar with fuzzy search on addresses (`module.app.aws_s3_bucket.assets`)  
- Snapshot compare: diff two states and highlight adds/changes/destroys

---

## Guidelines (for project contributors)

1. **Conventional commits**: `feat:`, `fix:`, `refactor:`, `docs:`, `ci:`, `build:`, `test:`  
2. Keep the service **stateless**; use object storage for optional snapshots  
3. Use **IaC** for deployment (Terraform, Pulumi, or Helm)  
4. Implement **CI/CD** with automated tests and image builds  
5. Enforce **least privilege** for any backend reads  
6. Provide clear **docs** for local dev, cloud deploy, and supported backends  
7. Add **redaction** options to mask sensitive attributes before rendering

---

## GitOps Alignment

- App and infra manifests reside in Git as the **single source of truth**  
- Deployments are **pull-based** using Argo CD or Flux, with continuous reconciliation  
- Optional: TerraScope can visualize GitOps state by:
  - Showing “synced/out-of-sync” status as a border color
  - Annotating nodes with the Git commit that produced the state
  - Emitting a small “drift” effect when the runtime diverges from Git

---

## Example Architecture

```
Terraform State (local upload or remote backend)
=> Parser / Normalizer (API)
=> Graph Builder (nodes, edges, metadata)
=> REST / WebSocket
=> Frontend (React/Svelte + Three.js/D3)
=> Interactive Constellation UI
```


---

## Data Sources

- `terraform.tfstate` JSON (local or remote backend)  
- Optional: `terraform plan -json` for planned changes  
- Optional: `terraform graph -json` for explicit dependency graph

---

## Possible Features

- 3D orbit simulation with physics inertia  
- Cost hints by resource type or tagging heuristics  
- Policy overlays (tagging required, encryption enabled)  
- Export screenshots or shareable links to a particular view/filter  
- Accessibility: keyboard navigation and high-contrast mode

---

## Deliverables

- Public GitHub repository containing:
  - Source code (frontend and backend)
  - IaC and CI/CD configuration
  - Setup documentation (local and cloud)
  - Example states for demo/testing
- Public URL for an interactive demo

---

> **Note**  
> These ideas are meant to inspire.  
> You can go minimal, educational, artistic, or deeply technical.  
> The objective is to make Terraform state understandable at a glance.
