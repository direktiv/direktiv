import UIPage from "~/pages/namespace/Explorer/Page";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/n/$namespace/explorer/page/$")({
  component: UIPage,
});
