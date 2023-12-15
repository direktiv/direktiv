import { InstantResposeFormSchema } from "./plugins/target/InstantResponse";
import { MethodsSchema } from "~/api/gateway/schema";
import { TargetFlowFormSchema } from "./plugins/target/TargetFlow";
import { z } from "zod";

export const EndpointFormSchema = z.object({
  direktiv_api: z.literal("endpoint/v1"),
  allow_anonymous: z.boolean().optional(),
  path: z.string().nonempty().optional(),
  timeout: z.number().int().positive().optional(),
  methods: z.array(MethodsSchema).nonempty().optional(),
  plugins: z
    .object({
      target: z.discriminatedUnion("type", [
        InstantResposeFormSchema,
        TargetFlowFormSchema,
      ]),
    })
    .optional(),
});

export type EndpointFormSchemaType = z.infer<typeof EndpointFormSchema>;
