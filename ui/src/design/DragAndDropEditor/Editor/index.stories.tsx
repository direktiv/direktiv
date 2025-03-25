import { DroppableElement, NonDroppableElement } from "../DroppableElement";
import type { Meta, StoryObj } from "@storybook/react";
import { Table, Text } from "lucide-react";
import { Dialog } from "~/design/Dialog";
import { DndContext } from "../Context.tsx";
import { DraggableElement } from "../DraggableElement";
import { DroppableSeparator } from "../DroppableSeparator/index.js";
import { EditModal } from "../EditModal";
import { useState } from "react";

const meta = {
  title: "Components/DragAndDropEditor",
  component: EditModal,
} satisfies Meta<typeof EditModal>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => <></>,
};

export const DragAndDrop = () => {
  const [hidden, setHidden] = useState<boolean>(false);
  const [name, setName] = useState<string>("Text");
  const [preview, setPreview] = useState<string>("Lorem ipsum...");
  const [dialogOpen, setDialogOpen] = useState<boolean>(false);

  const placeholder = {
    Table: "Header1  Cell1",
    Text: "Example Text",
  };

  const onMove = (element: string, target: string) => {
    if (target) {
      setName(element);

      element === "Table"
        ? setPreview(placeholder.Table)
        : setPreview(placeholder.Text);
    }
  };

  return (
    <DndContext onMove={onMove}>
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <div className="space-y-4">
          <h3>Draggable:</h3>
          <DraggableElement icon={Text} name="Text"></DraggableElement>
          <DraggableElement icon={Table} name="Table"></DraggableElement>

          <h3>Droppable:</h3>
          <div className="space-y-0">
            <NonDroppableElement
              name="Header"
              hidden={false}
              preview="This is the header"
            />
            <DroppableSeparator id="1" />
            <DroppableElement
              preview={preview}
              hidden={hidden}
              id="2"
              name={name}
              onHide={() => setHidden(!hidden)}
              setSelectedDialog={() => setDialogOpen(true)}
            />
            <DroppableSeparator id="3" />
            <NonDroppableElement
              name="Footer"
              hidden={false}
              preview="This is the footer"
            />
          </div>
        </div>
        <EditModal />
      </Dialog>
    </DndContext>
  );
};
