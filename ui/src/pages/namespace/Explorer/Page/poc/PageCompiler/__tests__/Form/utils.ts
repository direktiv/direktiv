import { HttpResponse, http } from "msw";

import { getUserDetailsResponse } from "../utils/api/samples";
import { setupServer } from "msw/node";
import { vi } from "vitest";

export const setupFormApi = () => {
  const apiRequestMock = vi.fn();
  const apiServer = setupServer(
    http.get("/user-details", () => HttpResponse.json(getUserDetailsResponse)),
    http.post("/save-user", (...args) => {
      apiRequestMock(...args);
      return HttpResponse.json({ status: "ok" });
    })
  );

  return { apiRequestMock, apiServer };
};
