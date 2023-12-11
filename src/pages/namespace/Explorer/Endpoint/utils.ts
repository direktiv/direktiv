import type { JSONSchema7Definition } from "json-schema";
import { RJSFSchema } from "@rjsf/utils";
import { endpointMethods } from "~/api/gateway/schema";
import { stringify } from "json-to-pretty-yaml";
import { useTranslation } from "react-i18next";

export const endpointHeader = {
  direktiv_api: "endpoint/v1",
};

export const defaultEndpointYaml = stringify(endpointHeader);

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
}: {
  name: string;
  pluginsObj: unknown;
}): JSONSchema7Definition => ({
  properties: {
    type: { enum: [name] },
    configuration: {},
  },
});

export const useEndpointFormSchema = (): RJSFSchema => {
  const { t } = useTranslation();

  return {
    properties: {
      method: {
        title: t("pages.explorer.tree.newEndpoint.form.method"),
        type: "string",
        // spread operator is required to convert from readonly to mutable array
        enum: [...endpointMethods],
      },
      plugins: {
        title: t("pages.explorer.tree.newEndpoint.form.plugins"),
        type: "array",
        items: {},
      },
    },
    required: ["method"],
    type: "object",
  };
};

export const addEndpointHeader = (endpointJSON: object) => ({
  ...endpointHeader,
  ...endpointJSON,
});
