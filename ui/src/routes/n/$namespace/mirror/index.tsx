import MirrorDetail from "~/pages/namespace/Mirror/Detail";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/n/$namespace/mirror/")({
  component: MirrorDetail,
});
