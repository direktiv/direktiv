export const validationsKeys = {
  validationsList: ({ hash }: { hash: string }) =>
    [
      {
        hash,
        scope: "validation-list",
      },
    ] as const,
};
