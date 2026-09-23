import { createFileRoute } from "@tanstack/react-router";
import { ProductWorkspacePage } from "@/components/ProductWorkspacePage";

export const Route = createFileRoute("/scheduled")({
  component: () => <ProductWorkspacePage kind="scheduled" />,
});
