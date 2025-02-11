import TokensPage from "~/pages/namespace/Permissions/Tokens";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/n/$namespace/permissions/tokens")({
  component: TokensPage,
});
