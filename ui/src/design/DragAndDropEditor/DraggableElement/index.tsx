import { FC, PropsWithChildren } from "react";
import { GripVertical, LucideIcon } from "lucide-react";

import Button from "~/design/Button";
import { useDraggable } from "@dnd-kit/core";

export type DraggableProps = PropsWithChildren & {
  name: string;
  icon: LucideIcon;
};

export const DraggableElement: FC<DraggableProps> = ({ name, icon: Icon }) => {
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
        className="absolute z-20 flex  h-full items-center opacity-50 group-hover:opacity-100"
      >
        <div className="flex w-5 items-center justify-center border-r rounded rounded-e-none hover:cursor-move hover:bg-gray-1 dark:hover:bg-gray-dark-1 ">
          <GripVertical />
        </div>
      </div>
      <Button asChild variant="outline" size="lg">
        <div className="bg-white dark:bg-black">
          <Icon size={16} />
          {name}
        </div>
      </Button>
    </div>
  );
};
