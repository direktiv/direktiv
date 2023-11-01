import { RJSFSchema } from "@rjsf/utils";
import { stringify } from "json-to-pretty-yaml";

export const serviceFormSchema: RJSFSchema = {
  properties: {
    image: {
      title: "Image",
      type: "string",
    },
    scale: {
      title: "Scale",
      type: "integer",
      enum: [0, 1, 2, 3, 4, 5, 6, 7, 8, 9],
    },
    size: {
      title: "size",
      type: "integer",
      enum: ["large", "medium", "small"],
    },
    cmd: {
      title: "Cmd",
      type: "string",
    },
  },
  required: ["image", "name"],
  type: "object",
};

export const serviceHeader = {
  direktiv_api: "service/v1",
};

export const addServiceHeader = (serviceJSON: object) => ({
  ...serviceHeader,
  ...serviceJSON,
});

export const defaultServiceYaml = stringify(serviceHeader);
