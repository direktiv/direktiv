import { Divide, Heading1 } from "lucide-react";
import { DraggableElementAdd, DraggableElementSort } from "./DraggableElement";

import { Card } from "../Card";
import { DndContext } from ".";
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
  const [actions, setActions] = useState<string[]>([]);

  return (
    <DndContext
      onDrop={(payload) => {
        if (payload.type === "add") {
          setActions((old) => [
            ...old,
            `🫳 you just added a ${payload.block.type}`,
          ]);
        }
        if (payload.type === "move") {
          setActions((old) => [
            ...old,
            `🫳 you just moved a ${payload.block.type} from ${payload.originPath.join(",")}`,
          ]);
        }
      }}
    >
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
      <Card className="mt-3 h-[100px] overflow-y-scroll p-5">
        {actions.reverse().map((action, index) => (
          <div key={index}>{action}</div>
        ))}
      </Card>
    </DndContext>
  );
};
