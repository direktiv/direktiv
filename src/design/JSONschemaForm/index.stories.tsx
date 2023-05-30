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
  },
};
export const Default = () =>
  <JSONSchemaForm schema={defSchema} />

// export const StateSchemaNoopForm =
//   <JSONSchemaForm schema={StateSchemaNoop} />
