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
  },
  required: ["method"],
  type: "object",
};

export const addEndpointHeader = (endpointJSON: object) => ({
  ...endpointHeader,
  ...endpointJSON,
});
