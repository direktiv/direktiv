export const validationKeys = {
  messagesList: ({ hash }: { hash: string }) =>
    [
      {
        hash,
        scope: "validation-list",
      },
    ] as const,
};
