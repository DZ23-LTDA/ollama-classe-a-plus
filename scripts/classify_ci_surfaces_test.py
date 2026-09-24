#!/usr/bin/env python3
"""Testes da classificacao de superficies de CI (scripts/classify-ci-surfaces.py).

Roda sem dependencias externas:
  python scripts/classify_ci_surfaces_test.py
"""
import importlib.util
import os
import unittest

HERE = os.path.dirname(os.path.abspath(__file__))
spec = importlib.util.spec_from_file_location(
    "classify_ci_surfaces", os.path.join(HERE, "classify-ci-surfaces.py")
)
mod = importlib.util.module_from_spec(spec)
spec.loader.exec_module(mod)


class ClassifyTest(unittest.TestCase):
    def req(self, *paths):
        return mod.classify(list(paths))["require_agentic_gates"]

    def docs_only(self, *paths):
        return mod.classify(list(paths))["docs_only"]

    def cats(self, path):
        return set(mod.categorize(path))

    # A) README-only -> docs-only, sem gates.
    def test_readme_only_is_docs_only(self):
        self.assertTrue(self.docs_only("README.md"))
        self.assertFalse(self.req("README.md"))
        self.assertEqual(self.cats("README.md"), {"DOCS_ONLY"})
        self.assertEqual(self.cats("docs/RELEASE_READINESS.md"), {"DOCS_ONLY"})

    # B) internal/agent/runtime.go -> agentic/core, exige gates.
    def test_agent_runtime_requires_gates(self):
        self.assertTrue(self.req("internal/agent/runtime.go"))
        self.assertFalse(self.docs_only("internal/agent/runtime.go"))
        self.assertIn("AGENTIC", self.cats("internal/agent/runtime.go"))

    # C) internal/multillm/router.go -> multi-provider/core, exige gates.
    def test_multillm_requires_gates(self):
        self.assertTrue(self.req("internal/multillm/router.go"))
        self.assertIn("MULTI_PROVIDER", self.cats("internal/multillm/router.go"))

    # D) server/routes.go -> core, exige gates.
    def test_server_routes_core(self):
        self.assertTrue(self.req("server/routes.go"))
        self.assertIn("CORE", self.cats("server/routes.go"))

    # E) app/ollama.iss -> build-release, exige gates.
    def test_iss_build_release(self):
        self.assertTrue(self.req("app/ollama.iss"))
        self.assertIn("BUILD_RELEASE", self.cats("app/ollama.iss"))

    # F) scripts/build_windows.ps1 -> build-release.
    def test_build_windows_build_release(self):
        self.assertTrue(self.req("scripts/build_windows.ps1"))
        self.assertIn("BUILD_RELEASE", self.cats("scripts/build_windows.ps1"))

    # G) pr-gate.yaml -> build-release.
    def test_pr_gate_workflow_build_release(self):
        self.assertTrue(self.req(".github/workflows/pr-gate.yaml"))
        self.assertIn("BUILD_RELEASE", self.cats(".github/workflows/pr-gate.yaml"))

    # H) CMakeLists.txt -> core/build.
    def test_cmake_core(self):
        self.assertTrue(self.req("CMakeLists.txt"))
        self.assertIn("CORE", self.cats("CMakeLists.txt"))

    # I) mlx file -> core/native.
    def test_mlx_core(self):
        self.assertTrue(self.req("mlx/backend.cpp"))
        self.assertIn("CORE", self.cats("mlx/backend.cpp"))

    # J) arquivo de codigo desconhecido -> FAIL SAFE (nao docs-only, exige gates).
    def test_unknown_code_fail_safe(self):
        self.assertTrue(self.req("weird/unknown_thing.xyz"))
        self.assertFalse(self.docs_only("weird/unknown_thing.xyz"))
        self.assertEqual(self.cats("weird/unknown_thing.xyz"), {"CORE"})

    # Mistura docs + codigo -> NAO docs-only, exige gates.
    def test_mixed_requires_gates(self):
        self.assertFalse(self.docs_only("README.md", "internal/agent/x.go"))
        self.assertTrue(self.req("README.md", "internal/agent/x.go"))

    # Conjunto vazio nao e docs-only (nada a pular; fail-safe conservador).
    def test_empty_not_docs_only(self):
        self.assertFalse(self.docs_only())


class ConsistencyTest(unittest.TestCase):
    """O paths-ignore do dz23-agentic-quality deve casar os padroes DOCS_ONLY."""

    def test_workflow_paths_ignore_matches_docs_only(self):
        import re

        wf = os.path.join(HERE, "..", ".github", "workflows", "dz23-agentic-quality.yaml")
        with open(wf, encoding="utf-8") as f:
            text = f.read()
        m = re.search(r"paths-ignore:\s*\n((?:\s*-\s*.+\n)+)", text)
        self.assertIsNotNone(m, "dz23-agentic-quality.yaml deve usar paths-ignore")
        listed = set(re.findall(r"-\s*'([^']+)'", m.group(1)))
        expected = set(mod.DOCS_ONLY_GLOBS)
        self.assertEqual(
            listed,
            expected,
            f"paths-ignore diverge dos DOCS_ONLY do classificador.\nworkflow={sorted(listed)}\nclassifier={sorted(expected)}",
        )


if __name__ == "__main__":
    unittest.main(verbosity=2)
