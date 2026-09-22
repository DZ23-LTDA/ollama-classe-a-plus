import { createFileRoute } from "@tanstack/react-router";
import { ProductWorkspacePage } from "@/components/ProductWorkspacePage";

export const Route = createFileRoute("/library")({
  component: () => <ProductWorkspacePage kind="library" />,
});
