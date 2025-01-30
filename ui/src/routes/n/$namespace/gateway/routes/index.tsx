import RoutesPage from "~/pages/namespace/Gateway/Routes";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/n/$namespace/gateway/routes/")({
  component: RoutesPage,
});
