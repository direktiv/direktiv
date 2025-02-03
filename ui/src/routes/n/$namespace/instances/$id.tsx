import InstancesDetail from "~/pages/namespace/Instances/Detail/";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/n/$namespace/instances/$id")({
  component: InstancesDetail,
});
