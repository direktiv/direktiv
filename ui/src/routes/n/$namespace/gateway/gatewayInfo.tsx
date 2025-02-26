import InfoPage from "~/pages/namespace/Gateway/Info";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/n/$namespace/gateway/gatewayInfo")({
  component: InfoPage,
});
