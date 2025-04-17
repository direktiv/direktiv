import { createFileRoute, redirect } from "@tanstack/react-router";

export const Route = createFileRoute("/n/$namespace/events/")({
  beforeLoad: () => {
    throw redirect({
      to: "/n/$namespace/events/history",
      from: "/n/$namespace",
    });
  },
});
