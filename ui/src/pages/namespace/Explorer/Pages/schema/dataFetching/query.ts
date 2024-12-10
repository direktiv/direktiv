import { DynamicString } from "../primitives/dynamicString";
import { SearchParms } from "./request/searchParams";
import { z } from "zod";

export const Query = z.object({
  id: z.string().min(1),
  endpoint: DynamicString,
  searchParms: SearchParms,
});
