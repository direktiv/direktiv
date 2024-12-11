import { Blocks, BlocksType } from ".";
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
  blocks: BlocksType["all"][];
};

export const QueryProvider = z.object({
  type: z.literal("queryProvider"),
  query: Query,
  blocks: z.array(Blocks.all),
}) satisfies z.ZodType<QueryProviderType>;
