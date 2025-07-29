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

export type DropzoneStatus = "hidden" | "allowed" | "forbidden";

type DroppableProps = PropsWithChildren & {
  payload: DropPayloadSchemaType;
  validate?: (
    payload: DragPayloadSchemaType | null,
    targetPath: BlockPathType
  ) => DropzoneStatus;
};

export const Dropzone: FC<DroppableProps> = ({
  payload,
  validate = () => "allowed",
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

  const status = validate(draggedPayload, targetPath);

  const { setNodeRef, isOver } = useDroppable({
    disabled: status === "hidden" || status === "forbidden",
    id: payload.targetPath.join("-"),
    data: payload,
  });

  const isDragging = !!activeDraggable;
  const showDropIndicator = isDragging && status !== "hidden";

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
              className={
                status === "forbidden"
                  ? twMergeClsx(
                      "w-fit bg-danger-8 transition-all dark:bg-danger-dark-8",
                      isOver && "bg-danger-10 dark:bg-danger-dark-10"
                    )
                  : twMergeClsx(
                      "w-fit bg-info-7 transition-all dark:bg-info-dark-7",
                      isOver && "bg-info-10 dark:bg-info-dark-10"
                    )
              }
            >
              <PlusCircle size={16} />
            </Badge>
          </div>
        </div>
      )}
    </div>
  );
};
