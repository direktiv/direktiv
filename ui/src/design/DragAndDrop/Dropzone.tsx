import {
  DragPayloadSchema,
  DragPayloadSchemaType,
  DropPayloadSchemaType,
} from "./schema";
import { FC, PropsWithChildren, useEffect, useMemo } from "react";
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
}) => {
  const { active: activeDraggable } = useDndContext();

  useEffect(
    // big problem here, payload was "new" on every rerender
    () => console.log("payload updated to ", payload),
    [payload]
  );

  useEffect(
    // no problem here
    () => console.log("activeDraggable updated to ", activeDraggable),
    [activeDraggable]
  );

  const status = useMemo(() => {
    const { targetPath } = payload;

    const parsedDragPayload = DragPayloadSchema.safeParse(
      activeDraggable?.data.current
    );

    const draggedPayload = parsedDragPayload.success
      ? parsedDragPayload.data
      : null;

    console.log("memoized validate");
    return validate(draggedPayload, targetPath);
  }, [validate, payload, activeDraggable?.data]);

  const { setNodeRef, isOver } = useDroppable({
    disabled: status !== "allowed",
    id: payload.targetPath.join("-"),
    data: payload,
  });

  const isDragging = !!activeDraggable;
  const showPlusIndicator = isDragging && status === "allowed";

  if (status === "hidden") {
    return null;
  }

  return (
    <>
      <div
        ref={setNodeRef}
        className={twMergeClsx(
          "relative h-[4px] w-full justify-center rounded-lg p-0",
          status === "allowed"
            ? [
                isDragging && "bg-primary-100 dark:bg-primary-800",
                isOver && "bg-primary-600 dark:bg-primary-600",
              ]
            : [isDragging && "bg-gray-4 dark:bg-gray-dark-4"]
        )}
      >
        {showPlusIndicator && isOver && (
          <div className="absolute inset-0 flex flex-col items-center justify-center">
            <div className="z-10 flex flex-col">
              <Badge className="bg-primary-600 dark:bg-primary-600">
                <PlusCircle size={16} />
              </Badge>
            </div>
          </div>
        )}
      </div>
    </>
  );
};
