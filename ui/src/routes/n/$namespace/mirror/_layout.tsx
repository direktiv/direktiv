import MirrorPage from "~/pages/namespace/Mirror";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/n/$namespace/mirror/_layout")({
  component: MirrorPage,
});
