import type { Meta, StoryObj } from "@storybook/react";
import {
  arraySchemaSample,
  basicExample,
  customArraySample,
  exampleThatThrowsAnError,
  simpleSample,
} from "./jsonSchemaExamples";
import Button from "../Button";
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
  <JSONSchemaForm schema={arraySchemaSample} />
);

export const CustomArraySampleForm = () => (
  <JSONSchemaForm schema={customArraySample} />
);

export const SampleFormWithFileInput = () => (
  <JSONSchemaForm schema={simpleSample} />
);

export const SampleFormWithAnError = () => (
  <>
    <header className="mb-4 font-semibold text-gray-12 dark:text-gray-dark-12">
      Selecting anything from the select field will show an error
    </header>
    <JSONSchemaForm schema={exampleThatThrowsAnError}>
      <Button type="submit">Submit</Button>
    </JSONSchemaForm>
  </>
);
