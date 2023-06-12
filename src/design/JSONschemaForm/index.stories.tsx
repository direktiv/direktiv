import {
  ArraySchemaSample,
  CustomArraySample,
  SimpleSample,
} from "./jsonSchemaExamples";
import type { Meta, StoryObj } from "@storybook/react";
import { JSONSchemaForm } from "../JSONschemaForm";
import { RJSFSchema } from "@rjsf/utils";

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
    occupation: {
      type: "string",
      enum: ["foo", "bar", "fuzz", "qux"],
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

export const ArraySchemaSampleForm = () => (
  <JSONSchemaForm schema={ArraySchemaSample} />
);

export const CustomArraySampleForm = () => (
  <JSONSchemaForm schema={CustomArraySample} />
);

export const SampleFormWithFileInput = () => (
  <JSONSchemaForm schema={SimpleSample} />
);
