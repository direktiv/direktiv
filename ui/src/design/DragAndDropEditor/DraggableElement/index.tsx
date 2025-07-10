import { FC, PropsWithChildren } from "react";

import { AllBlocksType } from "~/pages/namespace/Explorer/Page/poc/schema/blocks";
import { BlockPathType } from "~/pages/namespace/Explorer/Page/poc/PageCompiler/Block";
import { GripVertical } from "lucide-react";
import { useDraggable } from "@dnd-kit/core";

type DraggableProps = PropsWithChildren & {
  id: string;
  element: AllBlocksType;
  blockPath: BlockPathType | null;
};

export const DraggableElement: FC<DraggableProps> = ({
  id,
  element,
  children,
}) => {
  const data = element;
  const { attributes, listeners, setNodeRef, transform } = useDraggable({
    id,
    data,
  });

  const style = transform
    ? {
        transform: `translate3d(${transform.x}px, ${transform.y}px, 0) scale(1.05)`,
        zIndex: 20,
      }
    : {};

  return (
    <div style={style} className="relative">
      <div
        {...listeners}
        {...attributes}
        ref={setNodeRef}
        className="absolute z-0 -ml-4 flex h-full items-center p-0 text-gray-8 dark:text-gray-dark-8"
      >
        <div className="flex h-full w-5 items-center justify-center rounded rounded-e-none border-2 border-dashed border-gray-4 bg-white p-0 hover:cursor-move hover:border-solid hover:bg-gray-2 active:cursor-move active:border-solid active:bg-gray-2 dark:border-gray-dark-4 dark:bg-black dark:hover:bg-gray-dark-2">
          <GripVertical />
        </div>
      </div>
      {children}
    </div>
  );
};

type DraggableCreateProps = PropsWithChildren & {
  id: number;
  type: AllBlocksType["type"];
};

export const DraggableCreateElement: FC<DraggableCreateProps> = ({
  id,
  type,
  children,
}) => {
  const data = { type };
  const { attributes, listeners, setNodeRef, transform } = useDraggable({
    id,
    data,
  });

  const style = transform
    ? {
        transform: `translate3d(${transform.x}px, ${transform.y}px, 0) scale(1.05)`,
        zIndex: 50,
      }
    : {};

  return (
    <div style={style} className="relative">
      <div
        {...listeners}
        {...attributes}
        ref={setNodeRef}
        className="relative z-20 cursor-move"
      >
        {children}
      </div>
    </div>
  );
};
