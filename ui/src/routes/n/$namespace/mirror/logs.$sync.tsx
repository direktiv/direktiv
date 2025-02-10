import SyncDetail from "~/pages/namespace/Mirror/Detail/Sync/SyncDetail";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/n/$namespace/mirror/logs/$sync")({
  component: SyncDetail,
});
