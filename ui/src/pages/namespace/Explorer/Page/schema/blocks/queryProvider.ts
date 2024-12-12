import { Query, QueryType } from "../procedures/query";
import { AllBlocks } from ".";
import { AllBlocksType } from "./types";
import { z } from "zod";

/**
 * ⚠️ NOTE:
 * The type and the schema must be kept in sync to ensure 100% type safety.
 * It is currently possible to extend the schema without updating the type.
 * The schema needs to get the type input to avoid circular dependencies.
 */
export type QueryProviderType = {
  type: "query-provider";
  query: QueryType;
  blocks: AllBlocksType[];
};

export const QueryProvider = z.object({
  type: z.literal("query-provider"),
  query: Query,
  blocks: z.array(z.lazy(() => AllBlocks)),
}) satisfies z.ZodType<QueryProviderType>;
