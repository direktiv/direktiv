import ServicesPage from "~/pages/namespace/Services";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/n/$namespace/services/_layout")({
  component: ServicesPage,
});
