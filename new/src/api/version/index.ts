export const versionKeys = {
  all: (apiKey: string | undefined) => [{ scope: "versions", apiKey }] as const,
};
