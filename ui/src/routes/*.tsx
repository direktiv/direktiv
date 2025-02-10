import { ApiErrorSchemaType } from "~/api/errorHandling";
import ErrorPage from "~/util/router/ErrorPage";
import { createFileRoute } from "@tanstack/react-router";

const UnmatchedRoute = () => {
  const error: Error & ApiErrorSchemaType = {
    name: "invalid route",
    message: "invalid route",
    status: 404,
  };

  return <ErrorPage error={error}></ErrorPage>;
};

export const Route = createFileRoute("/*")({
  component: UnmatchedRoute,
});
