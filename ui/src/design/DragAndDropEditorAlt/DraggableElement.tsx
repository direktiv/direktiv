import { FC, PropsWithChildren } from "react";

import { GripVertical } from "lucide-react";
import { PayloadSchemaType } from "./schema";
import { useDraggable } from "@dnd-kit/core";

type DraggableProps = PropsWithChildren & {
  payload: PayloadSchemaType;
};

export const DraggableElement: FC<DraggableProps> = ({ payload, children }) => {
  const { attributes, listeners, setNodeRef, transform } = useDraggable({
    // TODO: use a better id?
    id: JSON.stringify(payload),
    data: payload,
  });

  const style = transform
    ? {
        transform: `translate3d(${transform.x}px, ${transform.y}px, 0) scale(1.05)`,
        zIndex: 20,
      }
    : {};

  return (
    <div style={style} className="relative m-1">
      <div
        {...listeners}
        {...attributes}
        ref={setNodeRef}
        className="absolute z-20 h-full text-gray-8 dark:text-gray-dark-8"
      >
        <div className="flex h-full w-5 items-center justify-center rounded rounded-e-none border-2 border-r-0 border-gray-4 bg-white p-0 hover:cursor-move hover:border-solid hover:bg-gray-2 active:cursor-move active:border-solid active:bg-gray-2 dark:border-gray-dark-4 dark:bg-black dark:hover:bg-gray-dark-2">
          <GripVertical />
        </div>
      </div>
      <div className="flex justify-center">
        <span className="mr-5"></span>
        <div className="w-full">{children}</div>
      </div>
    </div>
  );
};
