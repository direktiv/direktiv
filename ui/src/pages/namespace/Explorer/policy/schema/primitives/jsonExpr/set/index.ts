import { strictSingleKeyObject } from "../utils";
import { z } from "zod";

// when { [1, 2, "something"] };
export const SetJsonExprSchema = (jsonExprSchema: z.ZodTypeAny) =>
  strictSingleKeyObject("Set", z.array(jsonExprSchema));

type SetJsonExprSchemaType = z.infer<
  ReturnType<typeof SetJsonExprSchema>
>;
