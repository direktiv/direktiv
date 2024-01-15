export const permissionKeys = {
  get: ({ apiKey }: { apiKey?: string }) =>
    [{ scope: "permission", apiKey }] as const,
};
