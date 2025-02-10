import EventsPage from "~/pages/namespace/Events";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/n/$namespace/events")({
  component: EventsPage,
});
