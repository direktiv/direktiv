import { VersionSchema } from "./schema";
import { apiFactory } from "../utils";

export const getVersion = apiFactory({
  path: `/api/version`,
  method: "GET",
  schema: VersionSchema,
});
