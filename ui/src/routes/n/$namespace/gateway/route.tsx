import GatewayPage from "~/pages/namespace/Gateway";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/n/$namespace/gateway")({
  component: GatewayPage,
});
