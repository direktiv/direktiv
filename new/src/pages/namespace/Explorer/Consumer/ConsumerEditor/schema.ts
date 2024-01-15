import { z } from "zod";

export const ConsumerFormSchema = z.object({
  direktiv_api: z.literal("consumer/v1"),
  username: z.string().nonempty().optional(),
  password: z.string().nonempty().optional(),
  api_key: z.string().nonempty().optional(),
  tags: z.array(z.string()).optional(),
  groups: z.array(z.string()).optional(),
});

export type ConsumerFormSchemaType = z.infer<typeof ConsumerFormSchema>;
