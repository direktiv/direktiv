import { FC, PropsWithChildren } from "react";

import { GripVertical } from "lucide-react";
import { useDraggable } from "@dnd-kit/core";

export type DraggableProps = PropsWithChildren & {
  name: string;
};

export const DraggableElement: FC<DraggableProps> = ({ name, children }) => {
  const id = name;
  const { attributes, listeners, setNodeRef, transform } = useDraggable({
    id,
  });

  const style = transform
    ? {
        transform: `translate3d(${transform.x}px, ${transform.y}px, 0) scale(1.1)`,
        zIndex: 20,
      }
    : {};

  return (
    <div style={style} className="group relative pr-2">
      <div
        {...listeners}
        {...attributes}
        ref={setNodeRef}
        className="absolute z-20 flex h-full items-center opacity-50 group-hover:opacity-100"
      >
        <div className="flex w-5 items-center justify-center rounded rounded-e-none border-r hover:cursor-move hover:bg-gray-1 dark:hover:bg-gray-dark-1">
          <GripVertical />
        </div>
      </div>
      {children}
    </div>
  );
};
