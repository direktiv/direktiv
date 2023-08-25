export const authenticationKeys = {
  authentication: (apiKey: string | undefined) =>
    [{ scope: "authentication", apiKey }] as const,
};
