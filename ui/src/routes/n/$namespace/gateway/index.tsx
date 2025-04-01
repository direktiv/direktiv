import { createFileRoute, redirect } from "@tanstack/react-router";

export const Route = createFileRoute("/n/$namespace/gateway/")({
  beforeLoad: () => {
    throw redirect({
      to: "/n/$namespace/gateway/gatewayInfo",
      from: "/n/$namespace",
    });
  },
});
