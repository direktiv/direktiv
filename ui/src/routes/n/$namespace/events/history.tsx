import History from "~/pages/namespace/Events/History";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/n/$namespace/events/history")({
  component: History,
});
