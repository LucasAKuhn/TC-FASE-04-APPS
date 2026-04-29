# 🚩 ToggleMaster — Application Repository (TC-FASE-04-APPS)

> Código-fonte dos 5 microsserviços do sistema ToggleMaster, com pipelines CI/CD DevSecOps integrados via GitHub Actions.

**📦 Repositório GitOps (Manifestos K8s + Terraform):** [TC-FASE-04-GITOPS](https://github.com/julianopoklen/TC-FASE-04-GITOPS)

---

## 🎯 Arquitetura de Microsserviços

O sistema é composto por 5 microsserviços distintos, seguindo o padrão de **Persistência Poliglota**.

| Serviço | Tech Stack | Porta | Responsabilidade | Data Store |
| :--- | :--- | :--- | :--- | :--- |
| **Auth Service** | Go (Golang) | `8001` | Gerenciamento de autenticação e chaves de API. | **PostgreSQL** (`auth_db`) |
| **Flag Service** | Python | `8002` | CRUD de Feature Flags (Criação, Edição, Remoção). | **PostgreSQL** (`flag_db`) |
| **Targeting Service** | Python | `8003` | Gerenciamento de regras de segmentação complexas. | **PostgreSQL** (`targeting_db`) |
| **Evaluation Service** | Go (Golang) | `8004` | Avaliação de flags em tempo real (Alta performance). | **Redis** (Cache) + **SQS** (Eventos) |
| **Analytics Service** | Python | `8005` | Processamento assíncrono de eventos de avaliação. | **DynamoDB** (NoSQL) |

---

## 📂 Estrutura do Repositório

```
TC-FASE-04-APPS/
├── .github/workflows/       # Pipelines CI/CD — GitHub Actions (cross-repo push)
│   ├── auth-service.yaml
│   ├── flag-service.yaml
│   ├── targeting-service.yaml
│   ├── evaluation-service.yaml
│   └── analytics-service.yaml
├── auth-service/            # Código fonte — Autenticação (Go)
├── flag-service/            # Código fonte — Feature Flags (Python)
├── targeting-service/       # Código fonte — Segmentação (Python)
├── evaluation-service/      # Código fonte — Avaliação (Go)
├── analytics-service/       # Código fonte — Analytics (Python)
├── .gitignore
├── .trivyignore
└── README.md
```

---

## 🔄 Estratégia Multi-repo (GitOps)

Este repositório faz parte de uma arquitetura **Multi-repo** que separa fisicamente o código-fonte dos manifestos de infraestrutura:

```
TC-FASE-04-APPS (este repo)          TC-FASE-04-GITOPS
├── Código dos microsserviços         ├── k8s/ (manifestos K8s)
├── Workflows CI/CD                   ├── terraform/ (IaC)
└── Dockerfiles                       └── ArgoCD Application
         │                                     │
         │  ┌─────────────────────┐             │
         └──► CI faz cross-repo   ├─────────────┘
              push da nova tag
              no repo GitOps
```

### Por que separar?

1. **Evita loops de CI acidentais** — Commits em `k8s/` no monorepo re-disparavam workflows
2. **Governança de acessos** — Devs editam apps, SREs editam manifestos K8s
3. **Padrão corporativo** — Melhor prática de mercado para GitOps

---

## 🔐 CI/CD Pipeline (DevSecOps)

Cada serviço possui seu próprio workflow. O pipeline segue 3 etapas:

```
git push (código do serviço)
    └─► GitHub Actions (trigger path-based)
            ├─► Job 1: Build + SAST + Linter
            ├─► Job 2: SCA (Trivy FS scan)
            └─► Job 3: Docker Build + Container Scan + Push ECR
                    └─► Cross-repo update no TC-FASE-04-GITOPS
                            └─► ArgoCD detecta e faz deploy automático
```

| Workflow | Paths que disparam | Manifesto atualizado (no GitOps repo) |
| :--- | :--- | :--- |
| `auth-service.yaml` | `auth-service/**` | `k8s/01-auth.yaml` |
| `flag-service.yaml` | `flag-service/**` | `k8s/02-flag.yaml` |
| `targeting-service.yaml` | `targeting-service/**` | `k8s/03-targeting.yaml` |
| `evaluation-service.yaml` | `evaluation-service/**` | `k8s/04-evaluation.yaml` |
| `analytics-service.yaml` | `analytics-service/**` | `k8s/05-analytics.yaml` |

### Secrets necessárias no GitHub

| Secret | Descrição |
| :--- | :--- |
| `AWS_ACCESS_KEY_ID` | Credencial AWS Academy |
| `AWS_SECRET_ACCESS_KEY` | Credencial AWS Academy |
| `AWS_SESSION_TOKEN` | Token de sessão AWS Academy (4h) |
| `GITOPS_PAT` | **PAT do GitHub** com scope `repo` — usado para cross-repo push no TC-FASE-04-GITOPS |

### Variables necessárias

| Variable | Descrição |
| :--- | :--- |
| `AWS_REGION` | Região da AWS (ex: `us-east-1`) |

---

## 🚀 Quick Start (Desenvolvimento)

### Disparar CI/CD

Edite qualquer `temp.txt` dentro das pastas dos serviços e faça push:

```bash
git add .
git commit -m "trigger: force ci/cd for all services"
git push origin main
```

As 5 pipelines serão acionadas automaticamente. Após a conclusão, o ArgoCD detectará os novos manifestos no repo GitOps e fará deploy no cluster.

---

## 📋 Tech Challenge — Fase 4

**Projeto:** ToggleMaster — Observabilidade e Resiliência Ativa  
**Deadline:** 12/05/2026
