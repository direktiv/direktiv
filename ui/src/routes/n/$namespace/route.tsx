import Layout from "~/pages/namespace/Layout";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/n/$namespace")({
  component: Layout,
});
