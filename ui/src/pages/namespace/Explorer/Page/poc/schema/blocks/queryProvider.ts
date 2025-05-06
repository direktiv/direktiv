import { AllBlocks, AllBlocksType } from ".";
import { Query, QueryType } from "../procedures/query";
import { z } from "zod";

/**
 * ⚠️ NOTE:
 * The type and the schema must be kept in sync to ensure 100% type safety.
 * It is currently possible to extend the schema without updating the type.
 * The schema needs to get the type input to avoid circular dependencies.
 */
export type QueryProviderType = {
  type: "query-provider";
  queries: QueryType[];
  blocks: AllBlocksType[];
};

export const QueryProvider = z.object({
  type: z.literal("query-provider"),
  queries: z.array(Query),
  blocks: z.array(z.lazy(() => AllBlocks)),
}) satisfies z.ZodType<QueryProviderType>;
