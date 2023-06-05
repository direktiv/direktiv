import { RJSFSchema } from "@rjsf/utils";

export const ArraySchemaSample: RJSFSchema = {
  definitions: {
    Thing: {
      type: "object",
      properties: {
        name: {
          type: "string",
          default: "Default name",
        },
      },
    },
  },
  type: "object",
  properties: {
    listOfStrings: {
      type: "array",
      title: "A list of strings",
      items: {
        type: "string",
        default: "bazinga",
      },
    },
    ChoicesList: {
      type: "array",
      title: "A choices list",
      description: "select one",
      items: {
        type: "string",
        enum: ["foo", "bar", "fuzz", "qux"],
      },
      uniqueItems: false,
    },
  },
};

export const CustomArraySample: RJSFSchema = {
  title: "Custom array of strings",
  type: "array",
  items: {
    type: "string",
  },
};
export const SimpleSample: RJSFSchema = {
  title: "A registration form",
  description: "A simple form example.",
  type: "object",
  required: ["firstName", "lastName"],
  properties: {
    firstName: {
      type: "string",
      title: "First name",
      default: "Chuck",
      contentEncoding: "base64",
      contentMediaType: "image/png",
    },
    lastName: {
      type: "string",
      title: "Last name",
    },
    age: {
      type: "integer",
      title: "Age",
    },
    bio: {
      type: "string",
      title: "Bio",
      format: "data-url",
    },
    password: {
      type: "string",
      title: "Password",
      minLength: 3,
    },
    telephone: {
      type: "string",
      title: "Telephone",
      minLength: 10,
    },
  },
};

export const FunctionSchemaGlobal = {
  type: "object",
  required: ["id", "service"],
  properties: {
    id: {
      type: "string",
      title: "ID",
      description: "Function definition unique identifier.",
    },
    service: {
      type: "string",
      title: "Service",
      description: "The service being referenced.",
    },
  },
};

export const FunctionSchemaNamespace = {
  type: "object",
  required: ["id", "service"],
  properties: {
    id: {
      type: "string",
      title: "ID",
      description: "Function definition unique identifier.",
    },
    service: {
      type: "string",
      title: "Service",
      description: "The service being referenced.",
    },
  },
};

export const FunctionSchemaReusable = {
  type: "object",
  required: ["id", "image"],
  properties: {
    id: {
      type: "string",
      title: "ID",
      description: "Function definition unique identifier.",
    },
    image: {
      type: "string",
      title: "Image",
      description: "Image URI.",
      examples: [
        "direktiv/request",
        "direktiv/python",
        "direktiv/smtp-receiver",
        "direktiv/sql",
        "direktiv/image-watermark",
      ],
    },
    cmd: {
      type: "string",
      title: "CMD",
      description: "Command to run in container",
    },
    size: {
      type: "string",
      title: "Size",
      description: "Size of virtual machine",
    },
    scale: {
      type: "integer",
      title: "Scale",
      description: "Minimum number of instances",
    },
  },
};

export const FunctionSchemaSubflow = {
  type: "object",
  required: ["id", "workflow"],
  properties: {
    id: {
      type: "string",
      title: "ID",
      description: "Function definition unique identifier.",
    },
    workflow: {
      type: "string",
      title: "Workflow",
      description: "ID of workflow within the same namespace.",
    },
  },
};

export const FunctionSchema = {
  type: "object",
  required: ["type"],
  properties: {
    type: {
      enum: [
        "knative-workflow",
        "knative-namespace",
        "knative-global",
        "subflow",
      ],
      default: "knative-workflow",
      title: "Service Type",
      description: "Function type of new service",
    },
  },
  allOf: [
    {
      if: {
        properties: {
          type: {
            const: "knative-workflow",
          },
        },
      },
      then: FunctionSchemaReusable,
    },
    {
      if: {
        properties: {
          type: {
            const: "knative-namespace",
          },
        },
      },
      then: FunctionSchemaNamespace,
    },
    {
      if: {
        properties: {
          type: {
            const: "knative-global",
          },
        },
      },
      then: FunctionSchemaGlobal,
    },
    {
      if: {
        properties: {
          type: {
            const: "subflow",
          },
        },
      },
      then: FunctionSchemaSubflow,
    },
  ],
};
