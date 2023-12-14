import { MethodsSchema } from "~/api/gateway/schema";
import { stringify } from "json-to-pretty-yaml";
import yamljs from "js-yaml";
import { z } from "zod";

const InstantResposeFormSchema = z.object({
  type: z.literal("instant-response"),
  configuration: z.object({
    content_type: z.string().nonempty(),
    status_code: z.number().int().positive(),
    status_message: z.string().nonempty(),
  }),
});

const TargetFlowFormSchema = z.object({
  type: z.literal("target-flow"),
  configuration: z.object({
    flow: z.string().nonempty(),
    content_type: z.string().nonempty(),
    namespace: z.string().nonempty().optional(),
    async: z.boolean().optional(),
  }),
});

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

export const serializeEndpointFile = (yaml: string) => {
  let json;
  try {
    json = yamljs.load(yaml);
  } catch (e) {
    json = null;
  }

  const jsonParsed = EndpointFormSchema.safeParse(json);
  if (jsonParsed.success) {
    return jsonParsed.data;
  }
  return undefined;
};

const defaultEndpointFileJson: EndpointFormSchemaType = {
  direktiv_api: "endpoint/v1",
};

export const defaultEndpointFileYaml = stringify(defaultEndpointFileJson);
