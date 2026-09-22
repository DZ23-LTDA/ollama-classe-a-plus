import { createFileRoute } from "@tanstack/react-router";
import { CompanyWorkspacePage } from "@/components/CompanyWorkspacePage";

export const Route = createFileRoute("/company")({
  component: CompanyWorkspacePage,
});
