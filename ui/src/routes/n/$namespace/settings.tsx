import SettingsPage from "~/pages/namespace/Settings";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/n/$namespace/settings")({
  component: SettingsPage,
});
