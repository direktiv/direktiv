export const logKeys = {
  detail: (
    namespace: string,
    {
      apiKey,
      instance,
      route,
      activity,
      before,
      trace,
    }: {
      apiKey?: string;
      instance?: string;
      route?: string;
      activity?: string;
      before?: string;
      trace?: string;
    }
  ) =>
    [
      {
        scope: "logs",
        apiKey,
        namespace,
        instance,
        route,
        activity,
        before,
        trace,
      },
    ] as const,
};
