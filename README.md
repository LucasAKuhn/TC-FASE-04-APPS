# 🚩 ToggleMaster — Application Repository (TC-FASE-04-APPS)

> Código-fonte dos 5 microsserviços do sistema ToggleMaster, com pipelines CI/CD DevSecOps integrados via GitHub Actions.

**📦 Repositório GitOps (Manifestos K8s + Terraform):** [TC-FASE-04-GITOPS](https://github.com/LucasAKuhn/TC-FASE-04-GITOPS)

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

## 🔭 Observabilidade e Instrumentação (Fase 4)

Na Fase 4, todos os 5 microsserviços foram instrumentados para enviar dados de telemetria (traces, métricas e logs) via protocolo **OTLP** para um **OpenTelemetry Collector** centralizado, operando em modo gateway no namespace `monitoring` do cluster EKS.

### Estratégia de Instrumentação por Linguagem

A arquitetura poliglota do ToggleMaster exigiu abordagens distintas para cada runtime:

| Linguagem | Serviços | Abordagem |
| :--- | :--- | :--- |
| **Go** | `auth-service`, `evaluation-service` | Instrumentação **explícita** com `go.opentelemetry.io/otel`. Uso do interceptador HTTP `otelhttp.NewHandler` para captura automática de spans em cada requisição, e propagação de contexto via `context.Context` para Distributed Tracing. |
| **Python** | `flag-service`, `targeting-service`, `analytics-service` | **Auto-instrumentação** com `opentelemetry-instrumentation-flask`, `opentelemetry-instrumentation-requests` e `opentelemetry-instrumentation-psycopg2`/`botocore`. O SDK do OTel injeta spans automaticamente nos frameworks Flask, Requests e nos clientes de banco de dados. |

### Variáveis de Ambiente (Injetadas via K8s ConfigMap/Deployment)

Cada Deployment Kubernetes injeta as seguintes variáveis de ambiente do OpenTelemetry:

| Variável | Descrição |
| :--- | :--- |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | Endereço do OTel Collector (`otel-collector-opentelemetry-collector.monitoring.svc.cluster.local:4318`) |
| `OTEL_SERVICE_NAME` | Nome do serviço para identificação nos traces (ex: `auth-service`) |
| `OTEL_RESOURCE_ATTRIBUTES` | Atributos adicionais (ex: `service.namespace=toggle-master`) |

### Stack de Observabilidade (Provisionada no Repo GitOps)

Os dados de telemetria dos microsserviços são coletados e roteados pelo OTel Collector para os seguintes backends (todos provisionados via Terraform/Helm no repositório [TC-FASE-04-GITOPS](https://github.com/LucasAKuhn/TC-FASE-04-GITOPS)):

```
Microsserviços (OTLP) ──► OTel Collector (Gateway)
                              ├──► Métricas ──► Prometheus (remote write)
                              ├──► Logs     ──► Loki
                              └──► Traces   ──► New Relic (OTLP HTTP)
```

| Componente | Função |
| :--- | :--- |
| **Prometheus** | Armazenamento e consulta de métricas de infraestrutura e aplicação |
| **Loki** | Centralização e indexação de logs dos contêineres |
| **Promtail** | DaemonSet que coleta logs de todos os pods via filesystem dos nodes |
| **Grafana** | Visualização unificada — dashboards de métricas, logs e alertas |
| **New Relic** | APM comercial — Distributed Tracing, Service Map e análise de performance |

### Decisão Técnica: Por que New Relic?

O **New Relic** foi escolhido como ferramenta de APM pois aceita dados diretamente via protocolo OTLP (OpenTelemetry Protocol) **sem necessidade de agente proprietário**. Isso permitiu manter a arquitetura padronizada: o OTel Collector envia os traces diretamente para o endpoint `https://otlp.nr-data.net:4318`, sem instalar nenhum componente adicional nos pods.

### Alertas, Incidentes e Self-Healing

| Funcionalidade | Ferramenta | Descrição |
| :--- | :--- | :--- |
| **Alerta Inteligente** | Grafana Alerting | Regra de alerta configurada para disparar em cenários de degradação (ex: alta taxa de erros HTTP 5xx) |
| **Gestão de Incidentes** | PagerDuty | Integrado via Contact Point do Grafana (Events API v2) — incidentes são abertos automaticamente ao disparar alertas |
| **ChatOps** | Discord / Slack | Notificação automática com detalhes do alerta enviada para canal do time |
| **Self-Healing** | GitHub Actions | Workflow `self-healing.yml` acionado via `repository_dispatch` do Grafana — executa `kubectl rollout restart` automaticamente nos deployments do namespace `toggle-master` |

---

