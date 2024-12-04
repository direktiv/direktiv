import { FC, PropsWithChildren } from "react";

import { GripVertical } from "lucide-react";
import { useDraggable } from "@dnd-kit/core";

export type DraggableProps = PropsWithChildren & {
  name: string;
};

export const Draggable: FC<DraggableProps> = ({ children, name }) => {
  const id = name;
  const { attributes, listeners, setNodeRef, transform } = useDraggable({
    id,
  });

  const style = transform
    ? {
        transform: `translate3d(${transform.x}px, ${transform.y}px, 0) scale(1.1)`,
        zIndex: 1,
      }
    : {};

  return (
    <div style={style} className="group relative">
      <div
        {...listeners}
        {...attributes}
        ref={setNodeRef}
        className="absolute z-10 flex h-full items-center opacity-50 group-hover:opacity-100"
      >
        <div className="flex w-5 items-center justify-center border-r border-gray-400 hover:cursor-move hover:bg-gray-2">
          <GripVertical />
        </div>
      </div>
      {children}
    </div>
  );
};
