export const sessionKeys = {
  get: (apiKey?: string) => [{ scope: "session", apiKey }] as const,
};
