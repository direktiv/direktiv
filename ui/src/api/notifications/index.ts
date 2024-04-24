export const notificationKeys = {
  notifications: (namespace: string, { apiKey }: { apiKey?: string }) =>
    [
      {
        scope: "notifications",
        apiKey,
        namespace,
      },
    ] as const,
};
