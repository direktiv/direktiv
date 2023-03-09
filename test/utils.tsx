import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import React, { FC, PropsWithChildren } from "react";

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: false, // don't retry on error for tests
    },
  },
  logger: {
    // eslint-disable-next-line no-console
    log: console.log,
    warn: console.warn,
    error: () => null, // don't log errors for tests
  },
});

export const UseQueryWrapper: FC<PropsWithChildren> = ({ children }) => (
  <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
);
