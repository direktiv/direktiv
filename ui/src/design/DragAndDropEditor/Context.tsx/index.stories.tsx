import { AllBlocksType } from "~/pages/namespace/Explorer/Page/poc/schema/blocks";
import { BlockPathType } from "~/pages/namespace/Explorer/Page/poc/PageCompiler/Block";
import { DndContext } from ".";
import { DraggableElement } from "../DraggableElement";
import { DroppableSeparator } from "../DroppableSeparator";
import { HeadlineType } from "~/pages/namespace/Explorer/Page/poc/schema/blocks/headline";
import { useState } from "react";

export default {
  title: "Components/DragAndDropEditor/Context",
};

const initialBlocks: HeadlineType[] = [
  { type: "headline", level: "h3", label: "Headline 1" },
  { type: "headline", level: "h3", label: "Headline 2" },
  { type: "headline", level: "h3", label: "Headline 3" },
];

export const Default = () => {
  const [blocks, setBlocks] = useState<AllBlocksType[]>(initialBlocks);

  const moveBlock = (
    origin: BlockPathType,
    target: BlockPathType,
    block: AllBlocksType
  ) => {
    const fromIndex = origin[0] ?? 0;
    let toIndex = target[0] ?? 0;

    if (fromIndex < toIndex) toIndex -= 1;

    const updated = [...blocks];
    updated.splice(fromIndex, 1);
    updated.splice(toIndex, 0, block);
    setBlocks(updated);
  };

  const doSomething = () => 1;

  return (
    <DndContext onMove={moveBlock}>
      {blocks.map((block, index) => (
        <div key={index} className="my-2 flex flex-col items-center">
          <DroppableSeparator
            id={String(index)}
            position="before"
            blockPath={[index]}
            onDrop={() => doSomething}
          />
          <DraggableElement element={block} id={`${index}`} blockPath={[index]}>
            {block.type === "headline" && (
              <div className="border-2 border-dashed p-2">{block.label}</div>
            )}
          </DraggableElement>
        </div>
      ))}
    </DndContext>
  );
};
