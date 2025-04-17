import OpenapiDocPage from "~/pages/namespace/Gateway/OpenapiDoc";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/n/$namespace/gateway/openapiDoc")({
  component: OpenapiDocPage,
});
