import {
  DragPayloadSchema,
  DragPayloadSchemaType,
  DropPayloadSchemaType,
} from "./schema";
import { FC, PropsWithChildren } from "react";
import { useDndContext, useDroppable } from "@dnd-kit/core";

import Badge from "~/design/Badge";
import { PlusCircle } from "lucide-react";
import { twMergeClsx } from "~/util/helpers";

type DroppableProps = PropsWithChildren & {
  payload: DropPayloadSchemaType;
  isVisible?: (payload: DragPayloadSchemaType | null) => boolean;
};

export const Dropzone: FC<DroppableProps> = ({
  payload,
  isVisible = () => true,
  children,
}) => {
  const { active } = useDndContext();

  const parsedPayload = DragPayloadSchema.safeParse(active?.data.current);

  const isDisabled = !isVisible(
    parsedPayload.success ? parsedPayload.data : null
  );

  const { setNodeRef, isOver } = useDroppable({
    disabled: isDisabled,
    id: payload.targetPath.join("-"),
    data: payload,
  });

  const canDrop = !!active;

  return (
    <div
      ref={setNodeRef}
      className={twMergeClsx(
        "relative m-0 my-4 h-1 w-full justify-center rounded-lg p-0",
        isOver && "h-1 bg-gray-4 transition-all dark:bg-gray-dark-4",
        isDisabled && "invisible"
      )}
    >
      {children}
      {canDrop && (
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
