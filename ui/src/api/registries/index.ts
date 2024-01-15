export const registriesKeys = {
  registriesList: (namespace: string, { apiKey }: { apiKey?: string }) =>
    [
      {
        scope: "registries-list",
        apiKey,
        namespace,
      },
    ] as const,
};
