import { Blocks, BlocksType } from ".";
import {
  Query as DataQuery,
  QueryType as DataQueryType,
} from "../procedures/query";

import { z } from "zod";

/**
 * ⚠️ NOTE:
 * The type and the schema must be kept in sync to ensure 100% type safety.
 * It is currently possible to extend the schema without updating the type.
 * The schema needs to get the type input to avoid circular dependencies.
 */
export type QueryType = {
  type: "query";
  data: {
    id: string;
    query: DataQueryType;
    blocks: BlocksType["all"][];
  };
};

export const Query = z.object({
  type: z.literal("query"),
  data: z.object({
    id: z.string().min(1),
    query: DataQuery,
    blocks: z.array(Blocks.all),
  }),
}) satisfies z.ZodType<QueryType>;
