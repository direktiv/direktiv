export const policy = {
  get: (apiKey: string | undefined) => [{ scope: "policy", apiKey }] as const,
};
