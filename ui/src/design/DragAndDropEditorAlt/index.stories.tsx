import { DraggableElementAdd, DraggableElementSort } from "./DraggableElement";

import { Card } from "../Card";
import { DndContext } from ".";
import { Heading1 } from "lucide-react";
import { HeadlineType } from "~/pages/namespace/Explorer/Page/poc/schema/blocks/headline";
import { useState } from "react";

export default {
  title: "Components/DragAndDropEditorAlt",
};

const blocks: HeadlineType[] = [
  { type: "headline", level: "h3", label: "Headline 1" },
  { type: "headline", level: "h3", label: "Headline 2" },
  { type: "headline", level: "h3", label: "Headline 3" },
];

export const Default = () => {
  const [action, setAction] = useState("please drag something");

  return (
    <DndContext
      onDrop={(payload) => {
        setAction(
          `🫳 you just dropped a ${payload.block.type} (action: ${payload.type})`
        );
      }}
    >
      <Card className="mb-3 p-5">{action}</Card>
      <div className="flex gap-5">
        <Card className="w-[200px] p-3">
          <DraggableElementAdd
            payload={{
              type: "add",
              block: {
                type: "headline",
                label: "Headline",
                level: "h3",
              },
            }}
            icon={Heading1}
          >
            Headline
          </DraggableElementAdd>
        </Card>
        <Card className="grow p-3">
          {blocks.map((block, index) => {
            const blockPath = [index];
            return (
              <div key={index} className="my-2 flex flex-col items-center">
                {/* <DroppableSeparator
                        id={String(index)}
                        position="before"
                        blockPath={[index]}
                        onDrop={() => doSomething}
                        />*/}
                <DraggableElementSort
                  payload={{ type: "move", block, originPath: blockPath }}
                >
                  {block.type === "headline" && (
                    <div className="border-2 p-2">{block.label}</div>
                  )}
                </DraggableElementSort>
              </div>
            );
          })}
        </Card>
      </div>
    </DndContext>
  );
};
