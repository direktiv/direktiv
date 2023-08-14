import { z } from "zod";

const BooleanSchema = z.enum(["True", "False"]);

const ConditionSchema = z.object({
  name: z.enum(["ConfigurationsReady", "Ready", "RoutesReady"]),
  status: BooleanSchema,
  reason: z.string(),
  message: z.string(),
});

const ServiceSchema = z.object({
  info: z.object({
    name: z.string(),
    namespace: z.string(),
    workflow: z.string(),
    image: z.string(), // direktiv/request"
    cmd: z.string(),
    size: z.number(),
    minScale: z.number(),
    namespaceName: z.string(),
    path: z.string(),
    revision: z.string(),
    envs: z.object({}),
  }),
  status: BooleanSchema,
  conditions: z.array(ConditionSchema),
});

export const ServicesListSchema = z.object({
  config: z.object({
    maxscale: z.number(),
  }),
  functions: z.array(ServiceSchema),
});

export const ServiceFormSchema = z.object({
  cmd: z.string().nonempty(),
  image: z.string().nonempty(),
  minscale: z.number().int().gte(0).lte(3),
  scale: z.number().int().gte(1).lte(3),
});

export const ServiceDeletedSchema = z.null();

export const ServiceCreatedSchema = z.null();

export type ServicesListSchemaType = z.infer<typeof ServicesListSchema>;
export type ServiceFormSchemaType = z.infer<typeof ServiceFormSchema>;
