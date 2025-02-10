import PolicyPage from "~/pages/namespace/Permissions/Policy";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/n/$namespace/permissions/")({
  component: PolicyPage,
});
