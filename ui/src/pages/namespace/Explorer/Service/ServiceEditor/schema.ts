import { EnvironementVariableSchema } from "~/api/services/schema/services";
import { z } from "zod";

export const ServiceFormSchema = z.object({
  direktiv_api: z.literal("service/v1"),
  image: z.string().nonempty().optional(),
  scale: z.number().optional(), // further constraints on the number?
  size: z.string().nonempty().optional(), // should be an enum?
  cmd: z.string().nonempty().optional(),
  envs: z.array(EnvironementVariableSchema).nonempty().optional(), // should this be imported or is the form independent?
});

export type ServiceFormSchemaType = z.infer<typeof ServiceFormSchema>;
