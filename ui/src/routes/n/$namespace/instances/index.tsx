import InstancesListPage from "~/pages/namespace/Instances/List";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/n/$namespace/instances/")({
  component: InstancesListPage,
});
