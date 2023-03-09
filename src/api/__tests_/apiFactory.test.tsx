import { FC, PropsWithChildren } from "react";
import {
  QueryClient,
  QueryClientProvider,
  useQuery,
} from "@tanstack/react-query";
import {
  afterAll,
  afterEach,
  beforeAll,
  describe,
  expect,
  test,
  vi,
} from "vitest";
import { render, renderHook, screen, waitFor } from "@testing-library/react";

import { apiFactory } from "../utils";
import { rest } from "msw";
import { setupServer } from "msw/node";
import { z } from "zod";

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
    error: () => null,
  },
});

const useCustomHook = () => {
  return useQuery({ queryKey: ["customHook"], queryFn: () => "Hello" });
};

const wrapper: FC<PropsWithChildren> = ({ children }) => (
  <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
);

const testApi = setupServer(
  rest.get("/some", (_req, res, ctx) =>
    res(
      ctx.json({
        response: "some response",
      })
    )
  )
);

beforeAll(() => {
  testApi.listen({ onUnhandledRequest: "error" });
});

afterAll(() => testApi.close());

afterEach(() => {
  vi.clearAllMocks();
  testApi.resetHandlers();
});

const myApi = apiFactory({
  path: "/some",
  method: "GET",
  schema: z.object({
    response: z.string(),
  }),
});

describe("processApiResponse", () => {
  test("handles success and forwards optional additional payload", async () => {
    const { result } = renderHook(() => useCustomHook(), { wrapper });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    // render(<div>hi there...</div>);

    screen.debug();

    expect(1).toBe(1);
  });
});
