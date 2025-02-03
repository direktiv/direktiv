import { createFileRoute, redirect } from "@tanstack/react-router";

export const Route = createFileRoute("/n/$namespace/explorer/")({
  beforeLoad: (args) => {
    const { namespace } = args.params;

    throw redirect({
      to: "/n/$namespace/explorer/tree/$",
      params: { namespace, _splat: "/" },
    });
  },
});
