import GroupsPage from "~/pages/namespace/Permissions/Groups";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/n/$namespace/permissions/groups")({
  component: GroupsPage,
});
