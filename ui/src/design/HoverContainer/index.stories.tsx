import type { Meta, StoryObj } from "@storybook/react";

import { HoverContainer, HoverElement } from ".";
import { Card } from "../Card";

const meta = {
  title: "Components/HoverContainer",
  component: HoverContainer,
} satisfies Meta<typeof HoverContainer>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => (
    <HoverContainer id="id">
      <HoverElement>
        <p className="text-sm">visible on hover!</p>
      </HoverElement>

      <Card className="p-4">Hover over this card</Card>
    </HoverContainer>
  ),
};
