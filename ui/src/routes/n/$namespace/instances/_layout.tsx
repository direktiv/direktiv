import InstancesPage from "~/pages/namespace/Instances";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/n/$namespace/instances/_layout")({
  component: InstancesPage,
});
