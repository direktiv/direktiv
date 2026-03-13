import PolicyPage from "~/pages/namespace/Explorer/Policy";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/n/$namespace/explorer/policy/$")({
  component: PolicyPage,
});
