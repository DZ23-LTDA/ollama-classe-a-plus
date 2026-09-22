import { createFileRoute } from "@tanstack/react-router";
import { ProductWorkspacePage } from "@/components/ProductWorkspacePage";

export const Route = createFileRoute("/tasks")({
  component: () => <ProductWorkspacePage kind="tasks" />,
});
