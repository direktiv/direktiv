import { type Meta, type StoryObj } from "@storybook/react-vite";
import { ConditionsSchema } from "~/pages/namespace/Explorer/policy/schema/primitives/conditions";
import { PolicyConditionFlow } from "./PolicyConditionFlow";
import nestedBooleanGroups from "./policies/nestedBooleanGroups";
import whenAndUnless from "./policies/whenAndUnless";

const integrationClauses = ConditionsSchema.parse([
  ...nestedBooleanGroups,
  ...whenAndUnless,
]);

const meta = {
  title: "POC/Policy/Condition Flow",
  component: PolicyConditionFlow,
} satisfies Meta<typeof PolicyConditionFlow>;

export default meta;

type Story = StoryObj<typeof meta>;
export const Integration: Story = {
  args: {
    clauses: integrationClauses,
  },
};
