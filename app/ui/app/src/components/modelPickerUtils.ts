import type { Model } from "@/gotypes";

export function modelGroup(model: Model): string {
  if (model.kind === "router") return "Roteamento inteligente";
  if (model.kind === "remote") return model.provider || "APIs externas";
  return "Modelos locais";
}
