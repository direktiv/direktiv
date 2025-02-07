import ErrorPage from "~/util/router/ErrorPage";
import ExplorerWrapper from "~/pages/namespace/Explorer";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/n/$namespace/explorer")({
  component: ExplorerWrapper,
  errorComponent: ErrorPage,
});
