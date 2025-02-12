import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "~/design/Dropdown";
import {
  Eye,
  EyeOff,
  MoreVertical,
  PlusCircle,
  Settings,
  Trash2,
} from "lucide-react";
import { FC, PropsWithChildren } from "react";
import { HoverContainer, HoverElement } from "~/design/HoverContainer";
import { LogEntry, Logs } from "~/design/Logs";
import { useDndContext, useDroppable } from "@dnd-kit/core";

import Badge from "~/design/Badge";
import Button from "~/design/Button";
import { Card } from "~/design/Card";
import { DialogTrigger } from "~/design/Dialog";
import { twMergeClsx } from "~/util/helpers";

type DroppableProps = PropsWithChildren & {
  id: string;
  hidden: boolean;
  onHide?: () => void;
  name: string;
  preview: string;
  setSelectedDialog: (selectedDialog: dialogType) => void;
};

type dialogType = "edit" | "delete";

export const DroppableElement: FC<DroppableProps> = ({
  id,
  preview,
  setSelectedDialog,
  hidden,
  onHide,
  name,
}) => {
  const { setNodeRef, isOver } = useDroppable({
    id,
  });

  return (
    <div ref={setNodeRef} aria-label={name} className="relative">
      <HoverContainer>
        <Card
          noShadow
          className={twMergeClsx(
            "flex h-24 w-full items-center border-transparent justify-center bg-white dark:bg-black",
            isOver && " opacity-100",
            hidden && "bg-gray-2"
          )}
        >
          <HoverElement
            className={twMergeClsx(
              "bg-white dark:bg-black opacity-100",
              isOver && "hidden",
              hidden && "opacity-60"
            )}
            variant="alwaysVisibleLeft"
          >
            <Button icon variant="outline" onClick={onHide}>
              {hidden ? <EyeOff size={16} /> : <Eye size={16} />}
            </Button>
          </HoverElement>

          <HoverElement
            className={twMergeClsx(
              "bg-white dark:bg-black opacity-100",
              isOver && "hidden",
              hidden && "opacity-60"
            )}
            variant="alwaysVisibleRight"
          >
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button
                  variant="outline"
                  onClick={(e) => e.preventDefault()}
                  icon
                >
                  <MoreVertical size={16} />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent className="w-40">
                <DropdownMenuGroup>
                  <DialogTrigger
                    className="w-full"
                    onClick={() => {
                      setSelectedDialog("edit");
                    }}
                  >
                    <DropdownMenuItem>
                      <Settings className="mr-2 size-4" /> Edit
                    </DropdownMenuItem>
                  </DialogTrigger>
                  <DialogTrigger
                    className="w-full"
                    onClick={() => {
                      setSelectedDialog("delete");
                    }}
                  >
                    <DropdownMenuItem>
                      <Trash2 className="mr-2 size-4" /> Delete
                    </DropdownMenuItem>
                  </DialogTrigger>
                </DropdownMenuGroup>
              </DropdownMenuContent>
            </DropdownMenu>
          </HoverElement>

          <div className="flex flex-col">
            {isOver ? (
              <Badge className="bg-gray-10 ">
                <PlusCircle className="mr-2" size={16} />
                Replace this
              </Badge>
            ) : (
              <div
                className={twMergeClsx("opacity-100", hidden && "opacity-60")}
              >
                <div className="justify-center flex">
                  <Badge variant="outline" className="h-6">
                    {name}
                  </Badge>
                </div>
                <Logs>
                  <LogEntry className="pt-2 text-center">
                    <div className="truncate w-48">{preview}</div>
                  </LogEntry>
                </Logs>
              </div>
            )}
          </div>
        </Card>
      </HoverContainer>
    </div>
  );
};

type DropableProps = PropsWithChildren & {
  id: string;
};

export const Droppable3: FC<DropableProps> = ({ id, children }) => {
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
        "relative h-1 w-full bg-primary-200 transition-all",
        isOver && "h-12 bg-primary-500"
      )}
    >
      {children}
      {canDrop && <DropZone id={id} isOver={isOver} />}
    </div>
  );
};

const DropZone = ({ isOver, id }: { isOver: boolean; id: string }) => (
  <div className="absolute h-6 inset-0 flex flex-col items-center justify-center border-b-8">
    <div className="flex flex-col">
      <Badge variant="outline">
        {isOver ? (
          <>
            <PlusCircle className="mr-2" size={16} />
            {id}
          </>
        ) : (
          <>{!id ? "empty" : id}</>
        )}
      </Badge>
    </div>
  </div>
);

type NonDroppableElementProps = PropsWithChildren & {
  name: string;
  onHide?: () => void;
  onEdit?: () => void;
  hidden: boolean;
  preview: string;
};

export const NonDroppableElement: FC<NonDroppableElementProps> = ({
  name,
  onHide,
  onEdit,
  hidden,
  preview,
}) => (
  <div aria-label={name} className="relative">
    <HoverContainer>
      <Card
        noShadow
        className={twMergeClsx(
          "flex h-24 w-full items-center flex-col border-transparent justify-center bg-white dark:bg-black",
          hidden && "bg-gray-2 opacity-60"
        )}
      >
        <HoverElement
          className="bg-white dark:bg-black opacity-100"
          variant="alwaysVisibleLeft"
        >
          <Button icon variant="outline" onClick={onHide}>
            {hidden ? <EyeOff size={16} /> : <Eye size={16} />}
          </Button>
        </HoverElement>

        <HoverElement
          className="bg-white dark:bg-black opacity-100"
          variant="alwaysVisibleRight"
        >
          <Button icon variant="outline" onClick={onEdit}>
            <Settings size={16} />
          </Button>
        </HoverElement>

        <div className="justify-center flex ">
          <Badge variant="outline" className="h-6">
            {name}
          </Badge>
        </div>

        <Logs className="pt-2">
          <LogEntry>{preview}</LogEntry>
        </Logs>
      </Card>
    </HoverContainer>
  </div>
);
