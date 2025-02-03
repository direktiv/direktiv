import MonitoringPage from "~/pages/namespace/Monitoring";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/n/$namespace/monitoring")({
  component: MonitoringPage,
});
