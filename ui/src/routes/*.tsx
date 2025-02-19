import NotFoundPage from "~/util/router/NotFoundPage";
import { createFileRoute } from "@tanstack/react-router";

const UnmatchedRoute = () => <NotFoundPage />;

export const Route = createFileRoute("/*")({
  component: UnmatchedRoute,
});
