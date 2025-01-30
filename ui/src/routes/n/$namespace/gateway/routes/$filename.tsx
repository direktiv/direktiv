import RoutesDetailPage from "~/pages/namespace/Gateway/Routes/Detail";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/n/$namespace/gateway/routes/$filename")({
  component: RoutesDetailPage,
});
