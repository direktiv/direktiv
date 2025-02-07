import PermissionsPage from "~/pages/namespace/Permissions";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/n/$namespace/permissions")({
  component: PermissionsPage,
});
