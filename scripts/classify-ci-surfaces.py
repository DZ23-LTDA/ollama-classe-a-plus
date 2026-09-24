#!/usr/bin/env python3
"""Fonte unica e testavel da classificacao de superficies de CI.

Recebe caminhos alterados (args ou stdin, um por linha) e decide as categorias
e se os gates profundos (agentic) sao obrigatorios. Fail-safe: qualquer arquivo
de codigo/desconhecido conta como NAO docs-only -> exige CI. Em duvida, exige
mais CI, nunca menos.

Uso:
  git diff --name-only base..head | python scripts/classify-ci-surfaces.py
  python scripts/classify-ci-surfaces.py --require-agentic <files...>   # imprime true/false
  python scripts/classify-ci-surfaces.py --json <files...>

Este script e a fonte de verdade dos padroes DOCS_ONLY usados tambem pelo
`paths-ignore` de .github/workflows/dz23-agentic-quality.yaml (ver teste de
consistencia em scripts/classify_ci_surfaces_test.py).
"""
from __future__ import annotations

import fnmatch
import json
import sys

# Padroes considerados PURAMENTE editoriais (docs-only). Manter em sincronia
# com o paths-ignore do dz23-agentic-quality.yaml (validado por teste).
DOCS_ONLY_GLOBS = [
    "docs/**",
    "**/*.md",
    "**/*.mdx",
    "**/*.markdown",
    "LICENSE",
    "**/LICENSE",
    ".github/ISSUE_TEMPLATE/**",
    ".github/PULL_REQUEST_TEMPLATE.md",
    ".github/*.md",
]

# Categorias de codigo/build (qualquer uma => NAO docs-only => exige gates).
CATEGORY_GLOBS = {
    "AGENTIC": [
        "internal/agent/**",
        "server/agent_routes.go",
        "server/agent_*.go",
        "server/company_*.go",
        "server/plugin_routes.go",
        "server/companion_ws.go",
        "server/tel_agent_routes.go",
        "apps/mobile-agentic/**",
        "app/ui/app/**",
        "deploy/**",
    ],
    "MULTI_PROVIDER": [
        "internal/multillm/**",
        "internal/grok/**",
        "server/grok_routes.go",
        "server/cloud_proxy.go",
        "server/codex_proxy.go",
        "examples/dz23-*.json",
        "examples/*providers*.json",
    ],
    "WEB_MOBILE": [
        "app/ui/app/**",
        "apps/mobile-agentic/**",
    ],
    "BUILD_RELEASE": [
        ".github/workflows/**",
        ".github/rulesets/**",
        ".github/CODEOWNERS",
        "scripts/install*",
        "scripts/build*",
        "scripts/package_*",
        "scripts/classify-ci-surfaces.py",
        "app/ollama.iss",
        "app/ollama.rc",
        "SECURITY.md",
        "UPSTREAM_POLICY.md",
        "UPSTREAM_BASE_COMMIT",
    ],
    "CORE": [
        "**/*.go",
        "go.mod",
        "go.sum",
        "cmd/**",
        "server/**",
        "internal/**",
        "app/**",
        "CMakeLists.txt",
        "cmake/**",
        "llama/**",
        "ml/**",
        "mlx/**",
        "**/*.c",
        "**/*.cc",
        "**/*.cpp",
        "**/*.h",
        "**/*.hpp",
        "**/*.cu",
        "**/*.metal",
    ],
}

# Categorias que exigem os gates profundos agentic (dz23-agentic-quality).
AGENTIC_REQUIRING = {"AGENTIC", "WEB_MOBILE", "MULTI_PROVIDER", "CORE", "BUILD_RELEASE"}


def _match(path: str, globs: list[str]) -> bool:
    p = path.strip().replace("\\", "/")
    if not p:
        return False
    for g in globs:
        if fnmatch.fnmatch(p, g):
            return True
        # "**/x" casa qualquer diretorio INCLUSIVE nenhum (top-level).
        if g.startswith("**/") and fnmatch.fnmatch(p, g[3:]):
            return True
        # tambem casa o basename para padroes tipo "**/*.md".
        if g.startswith("**/") and fnmatch.fnmatch(p.rsplit("/", 1)[-1], g[3:]):
            return True
        # "dir/**" casa o proprio dir e tudo abaixo.
        if g.endswith("/**") and (p == g[:-3] or p.startswith(g[:-2])):
            return True
    return False


def is_docs_only_path(path: str) -> bool:
    return _match(path, DOCS_ONLY_GLOBS)


def categorize(path: str) -> set[str]:
    if is_docs_only_path(path):
        return {"DOCS_ONLY"}
    cats = {name for name, globs in CATEGORY_GLOBS.items() if _match(path, globs)}
    if not cats:
        # Fail-safe: arquivo nao-docs e nao-mapeado conta como CORE (exige CI).
        return {"CORE"}
    return cats


def classify(paths: list[str]) -> dict:
    paths = [p for p in (p.strip() for p in paths) if p]
    per_file = {p: sorted(categorize(p)) for p in paths}
    all_cats: set[str] = set()
    for cats in per_file.values():
        all_cats.update(cats)
    non_docs = any(cats != ["DOCS_ONLY"] for cats in per_file.values())
    require_agentic = any(
        (set(cats) & AGENTIC_REQUIRING) for cats in per_file.values()
    )
    return {
        "files": per_file,
        "categories": sorted(all_cats),
        "docs_only": (len(paths) > 0 and not non_docs),
        "require_agentic_gates": bool(require_agentic),
    }


def main(argv: list[str]) -> int:
    mode = "summary"
    args = list(argv)
    if args and args[0] in ("--require-agentic", "--json", "--docs-only"):
        mode = args.pop(0)
    paths = args if args else [ln for ln in sys.stdin.read().splitlines()]
    result = classify(paths)
    if mode == "--require-agentic":
        print("true" if result["require_agentic_gates"] else "false")
    elif mode == "--docs-only":
        print("true" if result["docs_only"] else "false")
    elif mode == "--json":
        print(json.dumps(result, indent=2, ensure_ascii=False))
    else:
        print(f"categories={','.join(result['categories']) or '(none)'}")
        print(f"docs_only={result['docs_only']}")
        print(f"require_agentic_gates={result['require_agentic_gates']}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv[1:]))
