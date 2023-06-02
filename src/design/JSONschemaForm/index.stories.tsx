import type { Meta, StoryObj } from "@storybook/react";
import { JSONSchemaForm } from "../JSONschemaForm";
import { RJSFSchema } from "@rjsf/utils";
import {
  FunctionSchema,
  FunctionSchemaNamespace,
  FunctionSchemaReusable,
  StateSchemaDelay,
} from "./jsonSchema";
const meta = {
  title: "Components/JSONSchemaForm",
  component: JSONSchemaForm,
} satisfies Meta<typeof JSONSchemaForm>;

export default meta;
type Story = StoryObj<typeof meta>;

const defSchema: RJSFSchema = {
  title: "A registration form",
  type: "object",
  required: ["firstName", "lastName"],
  properties: {
    password: {
      type: "string",
      title: "Password",
    },
    lastName: {
      type: "string",
      title: "Last name",
    },
    bio: {
      type: "string",
      title: "Bio",
    },
    firstName: {
      type: "string",
      title: "First name",
    },
    age: {
      type: "integer",
      title: "Age",
    },
  },
};

export const Default: Story = {
  render: ({ ...args }) => <JSONSchemaForm {...args} />,
  args: {
    schema: defSchema,
  },
  tags: ["autodocs"],
};

export const FunctionSchemaNamespaceForm = () => (
  <JSONSchemaForm schema={FunctionSchemaNamespace} />
);

export const FunctionSchemaReusableForm = () => (
  <JSONSchemaForm schema={FunctionSchemaReusable as RJSFSchema} />
);

export const FunctionSchemaSubflowForm = () => (
  <JSONSchemaForm schema={FunctionSchema as RJSFSchema} />
);

export const StateSchemaDelayForm = () => (
  <JSONSchemaForm schema={StateSchemaDelay as unknown as RJSFSchema} />
);
