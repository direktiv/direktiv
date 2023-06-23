import {
  ArraySchemaSample,
  CustomArraySample,
  SimpleSample,
  basicExample,
} from "./jsonSchemaExamples";
import type { Meta, StoryObj } from "@storybook/react";
import { JSONSchemaForm } from "../JSONschemaForm";

const meta = {
  title: "Components/JSONSchemaForm",
  component: JSONSchemaForm,
} satisfies Meta<typeof JSONSchemaForm>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: ({ ...args }) => <JSONSchemaForm {...args} />,
  args: {
    schema: basicExample,
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
