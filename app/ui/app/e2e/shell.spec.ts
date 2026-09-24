import { test, expect } from "@playwright/test";

// Smoke E2E do shell local-first (sem provider externo). Valida que o app
// carrega, mostra a home real e navega pelas rotas principais offline.
// Seletores por texto (robustos a mudanca de role/estrutura).
test.describe("Ollama Classe A+ shell", () => {
  test("home renders local-first composer and sidebar", async ({ page }) => {
    await page.goto("/");

    // Home real (redirect para o chat local): titulo central.
    await expect(
      page.getByText("O que posso fazer por você?", { exact: false }),
    ).toBeVisible();

    // Indicador de modo local-first (approvals/secrets protegidos).
    await expect(page.getByText(/Modo local-first/i)).toBeVisible();

    // Navegacao lateral principal presente.
    for (const item of ["Agente", "Tarefas", "Empresa"]) {
      await expect(page.getByText(item, { exact: false }).first()).toBeVisible();
    }
  });

  test("navigates to Tarefas (mission inbox) offline", async ({ page }) => {
    await page.goto("/");
    await page.getByText("Tarefas", { exact: false }).first().click();

    // Inbox de missoes renderiza sem depender de provider externo.
    await expect(
      page.getByText(/Inbox de missões|Veja missões/i).first(),
    ).toBeVisible();
  });

  // Jornada UI-first: navega direto pelas superficies principais offline e
  // confirma que cada shell renderiza (sem provider externo/modelo).
  const surfaces: Array<{ path: string; expect: RegExp }> = [
    { path: "/agentic", expect: /Mission Console|Agentic Runtime/i },
    { path: "/company", expect: /Company OS|Crie e opere sua empresa/i },
    { path: "/tasks", expect: /Inbox de missões|Veja missões|Tarefas/i },
  ];
  for (const s of surfaces) {
    test(`surface ${s.path} renders offline`, async ({ page }) => {
      await page.goto(s.path);
      await expect(page.getByText(s.expect).first()).toBeVisible();
      // O shell local-first permanece visivel em todas as rotas.
      await expect(page.getByText(/Modo local-first/i)).toBeVisible();
    });
  }
});
