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
  type: "queryProvider";
  query: QueryType;
  blocks: AllBlocksType[];
};

export const QueryProvider = z.object({
  type: z.literal("queryProvider"),
  query: Query,
  blocks: z.array(AllBlocks),
}) satisfies z.ZodType<QueryProviderType>;
