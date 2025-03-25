import type { Meta, StoryObj } from "@storybook/react";

import { Dialog } from "~/design/Dialog";
import { DroppableElement } from "./index";
import { EditModal } from "../EditModal";
import { useState } from "react";

const meta = {
  title: "Components/DragAndDropEditor/DroppableElement",
  component: DroppableElement,
} satisfies Meta<typeof DroppableElement>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: ({ ...args }) => (
    <Dialog>
      <DroppableElement {...args} />
    </Dialog>
  ),
  args: {
    preview: "Lorem ipsum dolor...",
    hidden: false,
    id: "1",
    name: "Text",
    setSelectedDialog: (selectedDropdown: string) =>
      alert("clicked " + selectedDropdown),
  },
  argTypes: {
    preview: {
      description: "Some info about the content of the element",
      control: "text",
      type: { name: "string", required: true },
    },
    hidden: {
      description: "Toggles the Visibility",
      control: "boolean",
      type: { name: "boolean", required: true },
    },
    id: {
      description: "The ID is needed for the drang and drop action",
      control: "text",
      type: { name: "string", required: true },
    },
    name: {
      description: "Name of the Element",
      control: "text",
      type: { name: "string", required: true },
    },
    setSelectedDialog: {
      description: "Needed if there is a DropDownMenu in the Button",
      control: "text",
      type: { name: "string", required: true },
    },
  },
};

export const ClickToHideOrEdit = () => {
  const [hidden, setHidden] = useState<boolean>(false);
  const [dialogOpen, setDialogOpen] = useState<boolean>(false);
  return (
    <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
      <DroppableElement
        preview="Lorem ipsum dolor..."
        hidden={hidden}
        id="1"
        name="Text"
        onHide={() => setHidden(!hidden)}
        setSelectedDialog={() => setDialogOpen(true)}
      />
      <EditModal />
    </Dialog>
  );
};
