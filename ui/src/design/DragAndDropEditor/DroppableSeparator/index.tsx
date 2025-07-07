import { FC, PropsWithChildren } from "react";
import { useDndContext, useDroppable } from "@dnd-kit/core";

import Badge from "~/design/Badge";
import { BlockPathType } from "~/pages/namespace/Explorer/Page/poc/PageCompiler/Block";
import { PlusCircle } from "lucide-react";
import { twMergeClsx } from "~/util/helpers";

type DroppableProps = PropsWithChildren & {
  id: string;
  blockPath: BlockPathType;
  position: "before" | "after" | undefined;
};

export const DroppableSeparator: FC<DroppableProps> = ({
  id,
  blockPath,
  children,
  position,
}) => {
  const { setNodeRef, isOver } = useDroppable({
    id,
  });

  const before = position === "before";
  const { active } = useDndContext();

  const canDrop = !!active;

  return (
    <div
      ref={setNodeRef}
      aria-label={id}
      className={twMergeClsx(
        "relative m-0 my-4 -ml-4 h-1 w-full justify-center rounded-lg p-0",
        before && "mb-4",
        isOver && "h-1 bg-gray-4 transition-all dark:bg-gray-dark-4"
      )}
    >
      {children}
      {canDrop && <DropZone id={id} isOver={isOver} blockPath={blockPath} />}
    </div>
  );
};

const DropZone = ({
  isOver,
}: {
  isOver: boolean;
  id: string;
  blockPath: BlockPathType;
}) => (
  <div className="absolute inset-0 flex flex-col items-center justify-center">
    <div className="z-10 flex flex-col">
      <Badge
        className={twMergeClsx(
          "w-fit bg-gray-8 transition-all dark:bg-gray-8",
          isOver && "bg-gray-10 dark:bg-gray-dark-10"
        )}
      >
        <PlusCircle size={16} />
      </Badge>
    </div>
  </div>
);
