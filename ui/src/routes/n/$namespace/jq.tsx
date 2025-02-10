import JqPlaygroundPage from "~/pages/namespace/JqPlayground";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/n/$namespace/jq")({
  component: JqPlaygroundPage,
});
