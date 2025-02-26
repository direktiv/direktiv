import OpenapiSpecificationPage from "~/pages/namespace/Explorer/OpenapiSpecification";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute(
  "/n/$namespace/explorer/openapiSpecification/$"
)({
  component: OpenapiSpecificationPage,
});
