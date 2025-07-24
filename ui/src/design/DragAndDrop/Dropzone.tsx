import {
  DragPayloadSchema,
  DragPayloadSchemaType,
  DropPayloadSchemaType,
} from "./schema";
import { FC, PropsWithChildren } from "react";
import { useDndContext, useDroppable } from "@dnd-kit/core";

import Badge from "~/design/Badge";
import { BlockPathType } from "~/pages/namespace/Explorer/Page/poc/PageCompiler/Block";
import { PlusCircle } from "lucide-react";
import { twMergeClsx } from "~/util/helpers";

type DroppableProps = PropsWithChildren & {
  payload: DropPayloadSchemaType;
  enable?: (
    payload: DragPayloadSchemaType | null,
    targetPath: BlockPathType
  ) => boolean;
};

export const Dropzone: FC<DroppableProps> = ({
  payload,
  enable = () => true,
  children,
}) => {
  const { active: activeDraggable } = useDndContext();
  const { targetPath } = payload;

  const parsedDragPayload = DragPayloadSchema.safeParse(
    activeDraggable?.data.current
  );

  const draggedPayload = parsedDragPayload.success
    ? parsedDragPayload.data
    : null;

  const isEnabled = enable(draggedPayload, targetPath);

  const { setNodeRef, isOver } = useDroppable({
    disabled: !isEnabled,
    id: payload.targetPath.join("-"),
    data: payload,
  });

  const isDragging = !!activeDraggable;
  const showDropIndicator = isDragging && isEnabled;

  return (
    <div
      ref={setNodeRef}
      className={twMergeClsx(
        "relative m-0 my-4 h-1 w-full justify-center rounded-lg p-0",
        isOver && "h-1 bg-gray-4 transition-all dark:bg-gray-dark-4"
      )}
    >
      {children}
      {showDropIndicator && (
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
      )}
    </div>
  );
};
