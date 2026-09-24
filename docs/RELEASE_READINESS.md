# Release Readiness — Ollama Classe A+

Este documento define a política de **um único SHA candidato passar por todo o
gate** antes de virar release. Ele não inventa checks: lista os workflows reais
já presentes em `.github/workflows/`.

## Princípio

Uma release só é válida quando **o mesmo commit** (o SHA candidato) passou por
todos os gates aplicáveis. Não vale juntar "parte verde na main" com "parte
verde num PR anterior". O RC é congelado num SHA e todos os workflows daquele
SHA devem estar verdes.

## Gates reais (contexts de status = nome do job)

| Workflow (`.github/workflows/`) | Job / context | Cobre |
|---|---|---|
| `class-a-plus-integrity.yaml` | **Preserve Class A+ surfaces** | guarda de integridade das superfícies + gofmt protegido |
| `dz23-agentic-quality.yaml` | **Go agentic and server gates** | go build/vet/test backend + server + browser operator |
| `dz23-agentic-quality.yaml` | **PostgreSQL RLS, Redis DLQ and OTLP integration** | RLS por org, DLQ/replay, OTLP |
| `dz23-agentic-quality.yaml` | **Web and mobile quality** | web typecheck/lint/tests/prod-audit + mobile typecheck/audit/policy |
| `release.yaml` | **Agentic release quality gate** | gate de qualidade agregado da release |
| `test.yaml` | test matrix (Linux/macOS/Windows) + race | testes Go multiplataforma |
| `test-install.yaml` | `test` | instala o fork de fonte e **verifica que o binário é o Classe A+** |
| `dz23-windows-installer.yaml` | `installer` | build CPU + Desktop + **instalador Windows (não assinado)** |
| `dz23-multi-provider.yaml` / `dz23-provider-smoke.yaml` | provider gates/smoke | roteamento multi-provider |

> GPU (CUDA/ROCm/Vulkan/MLX) permanece **manual / runner específico** — não é
> gate por commit. `release.yaml` cobre os builds pesados quando há runner.

## Gate automatizado (single-SHA)

O workflow [`release-readiness.yaml`](../.github/workflows/release-readiness.yaml)
**afirma** que o SHA candidato tem todos os checks reais acima verdes **naquele
commit** (não re-roda a matriz; verifica que ela já passou no SHA exato). Dispare
por `workflow_dispatch` (input `ref`) ou automaticamente ao criar uma tag `v*`.

## Procedimento de RC (um SHA)

1. Congele o candidato num SHA da `main` já verde.
2. Crie a tag do RC (versão nova; **não** reciclar `v0.1.0`). Ex.: `v0.2.0-rc.1`
   **apenas** se semanticamente apropriado após auditar o versionamento atual.
3. Confirme que **todos** os workflows acima executaram **naquele SHA** e estão
   verdes (`gh run list --commit <SHA>`), não em commits anteriores.
4. Gere e anexe os artefatos daquele SHA: instalador(es), **checksums SHA-256**,
   **SBOM**, metadata de build (versão + commit SHA), release notes verazes.
5. Marque a release como imutável e vinculada ao SHA testado.

## Estado verdadeiro (não marque PRODUCTION_READY sem prova)

- **Instalador Windows (amd64, não assinado):** produzido pelo CI
  (`dz23-windows-installer.yaml`). **Assinatura Authenticode = BLOCKED** (sem
  certificado). Uma release não deve aparecer como "signed" sem certificado real.
- **macOS/Linux packaging:** `release.yaml` cobre builds; empacotamento final
  assinado/notarizado depende de credenciais externas → **BLOCKED**.
- **Jornada E2E de primeira execução (install→onboarding→missão→artifact→
  approval→cancel→restart→persistência→upgrade→rollback):** parcialmente coberta
  por `test-install.yaml`; a jornada UI-first completa ainda precisa de smoke
  Playwright dedicado.

O estado agregado atual do produto é **RELEASE_CANDIDATE** (núcleo verde,
instalador Windows não assinado disponível), com assinatura e publicação
multiplataforma **BLOCKED_BY_EXTERNAL_DEPENDENCY**.
