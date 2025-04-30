import { FC, PropsWithChildren } from "react";
import { useDndContext, useDroppable } from "@dnd-kit/core";

import Badge from "~/design/Badge";
import { PlusCircle } from "lucide-react";
import { twMergeClsx } from "~/util/helpers";

type DroppableProps = PropsWithChildren & {
  id: string;
  position: "before" | "after" | undefined;
};

export const DroppableSeparator: FC<DroppableProps> = ({ id, children }) => {
  const { setNodeRef, isOver } = useDroppable({
    id,
  });

  const { active } = useDndContext();

  const canDrop = !!active;

  return (
    <div
      ref={setNodeRef}
      aria-label={id}
      className={twMergeClsx(
        "relative h-1 w-full justify-center rounded-lg bg-gray-4 dark:bg-gray-dark-4 transition-all",
        isOver && "h-1 bg-gray-8 dark:bg-gray-dark-8 transition-all"
      )}
    >
      {children}
      {canDrop && <DropZone id={id} isOver={isOver} />}
    </div>
  );
};

const DropZone = ({ isOver }: { isOver: boolean; id: string }) => (
  <div className="absolute inset-0 flex flex-col items-center justify-center ">
    <div className="flex flex-col z-10">
      <Badge
        className={twMergeClsx(
          "bg-gray-8 transition-all w-fit",
          isOver && "bg-gray-10"
        )}
      >
        <PlusCircle className="mr-2" size={16} />
        {isOver ? <>Insert here</> : <>...</>}
      </Badge>
    </div>
  </div>
);
