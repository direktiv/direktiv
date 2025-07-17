import { DndContext, DraggableElement } from ".";

import { AllBlocksType } from "~/pages/namespace/Explorer/Page/poc/schema/blocks";
import { BlockPathType } from "~/pages/namespace/Explorer/Page/poc/PageCompiler/Block";
import { Card } from "../Card";
import { HeadlineType } from "~/pages/namespace/Explorer/Page/poc/schema/blocks/headline";
import path from "path";
import { pathToId } from "~/pages/namespace/Explorer/Page/poc/PageCompiler/context/utils";
import { useState } from "react";

export default {
  title: "Components/DragAndDropEditorAlt",
};

const initialBlocks: HeadlineType[] = [
  { type: "headline", level: "h3", label: "Headline 1" },
  { type: "headline", level: "h3", label: "Headline 2" },
  { type: "headline", level: "h3", label: "Headline 3" },
];

export const Default = () => {
  const [blocks, setBlocks] = useState<AllBlocksType[]>(initialBlocks);

  return (
    <DndContext
      onDrop={(payload) => {
        alert(
          `🫳 you just dropped something! \n ${payload.type} a ${payload.block.type}`
        );
      }}
    >
      <div className="flex gap-5">
        <Card className="w-[200px] p-3">
          <DraggableElement
            payload={{
              type: "add",
              block: {
                type: "headline",
                label: "Headline",
                level: "h3",
              },
            }}
          >
            <div className="border-2 p-2">Headline</div>
          </DraggableElement>
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
                <DraggableElement
                  payload={{ type: "move", block, originPath: blockPath }}
                >
                  {block.type === "headline" && (
                    <div className="border-2 p-2">{block.label}</div>
                  )}
                </DraggableElement>
              </div>
            );
          })}
        </Card>
      </div>
    </DndContext>
  );
};
