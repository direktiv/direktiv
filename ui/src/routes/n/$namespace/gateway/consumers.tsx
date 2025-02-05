import ConsumersPage from "~/pages/namespace/Gateway/Consumers";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/n/$namespace/gateway/consumers")({
  component: ConsumersPage,
});
