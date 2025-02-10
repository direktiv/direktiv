import RolesPage from "~/pages/namespace/Permissions/Roles";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/n/$namespace/permissions/")({
  component: RolesPage,
});
