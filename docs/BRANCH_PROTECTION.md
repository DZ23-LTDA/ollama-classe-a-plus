# Proteção da branch `main` — Ollama Classe A+

> **Status: BLOCKED_BY_EXTERNAL_DEPENDENCY — GITHUB_REPOSITORY_ADMIN_PERMISSION**
>
> Aplicar proteção de branch/ruleset exige permissão de **admin** no
> repositório. Na última verificação, a permissão efetiva do executor era
> `admin:false, maintain:false`. O ruleset abaixo está pronto; **um
> administrador precisa aplicá-lo**.

## O que o ruleset garante

- PR obrigatório para mudar a `main` (1 aprovação, resolução de threads).
- Bloqueio de **force-push** (`non_fast_forward`) e de **delete** (`deletion`).
- **Required status checks** com os workflows reais do projeto, e a branch
  precisa estar atualizada antes do merge (`strict`).
- Dismiss de aprovação stale ao novo push.

Os contexts exigidos usam os **nomes reais** dos jobs (ver
[`RELEASE_READINESS.md`](RELEASE_READINESS.md)):
`Preserve Class A+ surfaces`, `Go agentic and server gates`,
`PostgreSQL RLS, Redis DLQ and OTLP integration`, `Web and mobile quality`.

## Como aplicar (admin do repositório)

Ruleset pronto em [`.github/rulesets/main-protection.json`](../.github/rulesets/main-protection.json).

```bash
gh api -X POST repos/DZ23-LTDA/ollama-classe-a-plus/rulesets \
  --input .github/rulesets/main-protection.json
```

Ou via UI: **Settings → Rules → Rulesets → New ruleset → Import** o JSON acima.

## Itens adicionais que dependem de admin

- Habilitar **Private Vulnerability Reporting** (Settings → Code security and
  analysis) — ver [`../SECURITY.md`](../SECURITY.md).
- Proteção de **tags de release** (ruleset com `target: tag`, `include:
  refs/tags/v*`, regra `deletion` + `non_fast_forward`).
- Exigir **commits/tags assinados** para releases (`required_signatures`), se a
  equipe adotar assinatura GPG/Sigstore.
- `CODEOWNERS`: pode ser adicionado quando os owners reais forem confirmados;
  então habilite `require_code_owner_review: true` no ruleset.

## Verificação após aplicar

```bash
gh api repos/DZ23-LTDA/ollama-classe-a-plus/rulesets            # lista rulesets
gh api repos/DZ23-LTDA/ollama-classe-a-plus/branches/main -q '.protected'
```
