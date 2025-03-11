import { createFileRoute, redirect } from "@tanstack/react-router";

export const Route = createFileRoute("/n/$namespace/permissions/")({
  beforeLoad: () => {
    throw redirect({
      to: "/n/$namespace/permissions/roles",
      from: "/n/$namespace",
    });
  },
});
