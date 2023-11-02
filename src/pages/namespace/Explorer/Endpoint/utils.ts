import {
  PluginJSONSchemaType,
  PluginsListSchemaType,
  endpointMethods,
} from "~/api/gateway/schema";

import type { JSONSchema7Definition } from "json-schema";
import { RJSFSchema } from "@rjsf/utils";

export const endpointHeader = {
  direktiv_api: "endpoint/v1",
};

/**
 * takes the plugins server response and returns an array of plugin names
 */
const getPluginsList = (pluginsObj: PluginsListSchemaType) =>
  Object.keys(pluginsObj.data);

/**
 * input:
 
 {
    "$defs": {
      "examplePluginConfig": {
        "additionalProperties": false,
        "properties": {
          "echo_value": {
            "type": "string"
          }
        },
        "required": [
          "echo_value"
        ],
        "type": "object"
      }
    },
    "$id": "https://github.com/direktiv/direktiv/pkg/refactor/gateway/example-plugin-config",
    "$ref": "#/$defs/examplePluginConfig",
    "$schema": "https://json-schema.org/draft/2020-12/schema"
  } 

  transformed output:
  {
    properties: {
      type: { enum: ["examplePluginConfig"] },
      configuration: {
        "additionalProperties": false,
        "properties": {
          "echo_value": {
            "type": "string"
          }
        },
        "required": [
          "echo_value"
        ],
        "type": "object"
      }
    },
  }

 */
export const generatePluginJSONSchema = ({
  name,
  pluginsObj,
}: {
  name: string;
  pluginsObj: PluginJSONSchemaType;
}): JSONSchema7Definition => ({
  properties: {
    type: { enum: [name] },
    configuration: Object.values(pluginsObj.$defs)?.[0] ?? {},
  },
});

export const endpointBaseFormSchema = (
  pluginsObj: PluginsListSchemaType
): RJSFSchema => {
  const pluginSchemas = Object.entries(pluginsObj.data).map(([name, value]) =>
    generatePluginJSONSchema({ name, pluginsObj: value })
  );

  return {
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
            type: { enum: getPluginsList(pluginsObj) },
          },
          required: ["type"],
          dependencies: {
            type: {
              oneOf: pluginSchemas,
            },
          },
        },
      },
    },
    required: ["method", "plugins"],
    type: "object",
  };
};

export const addEndpointHeader = (endpointJSON: object) => ({
  ...endpointHeader,
  ...endpointJSON,
});
