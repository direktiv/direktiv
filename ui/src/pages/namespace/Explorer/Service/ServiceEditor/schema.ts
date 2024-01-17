import { EnvironementVariableSchema } from "~/api/services/schema/services";
import { z } from "zod";

export const ServiceFormSchema = z.object({
  direktiv_api: z.literal("service/v1"),
  image: z.string().nonempty().optional(),
  scale: z.number().min(0).lt(10).optional(),
  size: z.string().optional(),
  cmd: z.string().optional(),
  envs: z.array(EnvironementVariableSchema).nonempty().optional(),
});

export type ServiceFormSchemaType = z.infer<typeof ServiceFormSchema>;
