import ServicePage from "~/pages/namespace/Explorer/Service";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/n/$namespace/explorer/service/$")({
  component: ServicePage,
});
