import ExplorerPage from "~/pages/namespace/Explorer/Tree";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/n/$namespace/explorer/tree/$")({
  component: ExplorerPage,
});
