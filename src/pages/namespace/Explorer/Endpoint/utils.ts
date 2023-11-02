import { RJSFSchema } from "@rjsf/utils";
import { endpointMethods } from "~/api/gateway/schema";

export const endpointHeader = {
  direktiv_api: "endpoint/v1",
};

export const endpointBaseFormSchema: RJSFSchema = {
  properties: {
    method: {
      title: "method",
      type: "integer",
      // spread operator is required to convert from readonly to mutable array
      enum: [...endpointMethods],
    },
    plugins: {
      title: "Plugins",
      type: "array",
      items: {
        type: "object",
        properties: {
          type: { enum: ["A", "B"] },
        },
        required: ["type"],
        dependencies: {
          type: {
            oneOf: [
              {
                properties: {
                  type: { enum: ["A"] },
                  configuration: {
                    type: "object",
                    properties: {
                      version: {
                        type: "integer",
                      },
                    },
                    required: ["version"],
                    additionalProperties: false,
                  },
                },
              },
              {
                properties: {
                  type: { enum: ["B"] },
                  configuration: {
                    type: "object",
                    properties: {
                      some: {
                        type: "integer",
                      },
                    },
                    required: ["some"],
                    additionalProperties: false,
                  },
                },
              },
            ],
          },
        },
      },
    },
  },
  required: ["method", "plugins"],
  type: "object",
};

export const addEndpointHeader = (endpointJSON: object) => ({
  ...endpointHeader,
  ...endpointJSON,
});
