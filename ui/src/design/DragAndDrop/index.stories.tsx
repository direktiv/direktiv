import { DragablePaletteItem, SortableItem } from "./Draggable";

import { Card } from "../Card";
import { DndContext } from ".";
import { Dropzone } from "./Dropzone";
import { Heading1 } from "lucide-react";
import { HeadlineType } from "~/pages/namespace/Explorer/Page/poc/schema/blocks/headline";
import { useState } from "react";

export default {
  title: "Components/DragAndDrop",
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
        const { drag, drop } = payload;
        if (drag.type === "add") {
          setActions((old) => [
            ...old,
            `you just added a ${drag.blockType} at ${drop.targetPath.join(",")}`,
          ]);
        }
        if (drag.type === "move") {
          setActions((old) => [
            ...old,
            `you just moved a ${drag.block.type} from ${drag.originPath.join(",")} to ${drop.targetPath.join(",")}`,
          ]);
        }
      }}
    >
      <div className="flex gap-5">
        <Card className="w-[200px] p-3">
          <DragablePaletteItem
            payload={{
              type: "add",
              blockType: "headline",
            }}
            icon={Heading1}
          >
            Headline
          </DragablePaletteItem>
        </Card>
        <Card className="grow p-3">
          {blocks.map((block, index) => {
            const blockPath = [index];
            return (
              <div key={index} className="my-2 flex flex-col items-center">
                <Dropzone payload={{ targetPath: [index] }} />
                <SortableItem
                  payload={{ type: "move", block, originPath: blockPath }}
                >
                  {block.type === "headline" && (
                    <div className="border-2 p-2">{block.label}</div>
                  )}
                </SortableItem>
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
