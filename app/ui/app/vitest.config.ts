import { defineConfig, mergeConfig } from "vite";
import path from "path";
import { configDefaults } from "vitest/config";
import baseConfig from "./vite.config";

export default defineConfig((configEnv) =>
  mergeConfig(
    baseConfig(configEnv),
    defineConfig({
      resolve: {
        alias: {
          "@": path.resolve(__dirname, "./src"),
          "@/gotypes": path.resolve(__dirname, "./codegen/gotypes.gen.ts"),
        },
      },
      test: {
        environment: "node",
        globals: true,
        // Os specs em e2e/ sao Playwright (rodados por `playwright test`, nao
        // pelo vitest). Sem este exclude, o glob padrao **/*.spec.ts coletaria
        // e2e/*.spec.ts e falharia ao importar @playwright/test.
        exclude: [...configDefaults.exclude, "e2e/**"],
      },
    }),
  ),
);
