export const logKeys = {
  detail: (
    namespace: string,
    {
      apiKey,
      instance,
      route,
      activity,
      trace,
    }: {
      apiKey?: string;
      instance?: string;
      route?: string;
      activity?: string;
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
        trace,
      },
    ] as const,
};
