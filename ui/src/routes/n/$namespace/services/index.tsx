import ServicesListPage from "~/pages/namespace/Services/List";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/n/$namespace/services/")({
  component: ServicesListPage,
});
