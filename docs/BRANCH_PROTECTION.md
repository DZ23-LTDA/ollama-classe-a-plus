# Proteção da branch `main` — Ollama Classe A+

> **Status: APLICADO (2026-09-24)** com a conta admin `DZ23-LTDA`.
>
> - `main-protection` [branch] — ruleset **active** (`main.protected = true`).
> - `release-tag-protection` [tag] — ruleset **active** (protege `refs/tags/v*`).
> - Private Vulnerability Reporting — **habilitado**.
>
> Os JSONs versionados em `.github/rulesets/` são a fonte de verdade; reaplique
> com os comandos abaixo se precisar recriar.

## O que o ruleset garante

- PR obrigatório para mudar a `main` (1 aprovação, resolução de threads).
- Bloqueio de **force-push** (`non_fast_forward`) e de **delete** (`deletion`).
- **Required status checks** com os workflows reais do projeto, e a branch
  precisa estar atualizada antes do merge (`strict`).
- Dismiss de aprovação stale ao novo push.

Os contexts exigidos sao **estaveis e sempre criados** em todo PR:
`Preserve Class A+ surfaces` (workflow `class-a-plus-integrity`, sem filtro de
paths) e `PR gate` (workflow `pr-gate`, agregador que roda em todo PR e, quando
o PR toca runtime/web/mobile, espera e verifica os gates profundos de
`dz23-agentic-quality`). Isso evita o deadlock de exigir checks com filtro de
paths que nao sao criados em PRs docs-only.

## Como aplicar (admin do repositório)

Ruleset pronto em [`.github/rulesets/main-protection.json`](../.github/rulesets/main-protection.json).

```bash
gh api -X POST repos/DZ23-LTDA/ollama-classe-a-plus/rulesets \
  --input .github/rulesets/main-protection.json
```

Ou via UI: **Settings → Rules → Rulesets → New ruleset → Import** o JSON acima.

## Itens adicionais

- **Private Vulnerability Reporting** — ✅ habilitado (ver [`../SECURITY.md`](../SECURITY.md)).
- Proteção de **tags de release** — ✅ aplicada via
  [`.github/rulesets/tag-protection.json`](../.github/rulesets/tag-protection.json).
- Exigir **commits/tags assinados** para releases (`required_signatures`): ainda
  não aplicado — depende de a equipe adotar assinatura GPG/Sigstore.
- `CODEOWNERS`: pode ser adicionado quando os owners reais forem confirmados;
  então habilite `require_code_owner_review: true` no ruleset.

## Verificação após aplicar

```bash
gh api repos/DZ23-LTDA/ollama-classe-a-plus/rulesets            # lista rulesets
gh api repos/DZ23-LTDA/ollama-classe-a-plus/branches/main -q '.protected'
```
