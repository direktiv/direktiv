import ConsumerPage from "~/pages/namespace/Explorer/Consumer";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/n/$namespace/explorer/consumer/$")({
  component: ConsumerPage,
});
