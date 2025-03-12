export const logKeys = {
  detail: (
    namespace: string,
    {
      apiKey,
      instance,
      route,
      activity,
      trace,
      last,
    }: {
      apiKey?: string;
      instance?: string;
      route?: string;
      activity?: string;
      trace?: string;
      last?: number;
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
        last,
      },
    ] as const,
};
