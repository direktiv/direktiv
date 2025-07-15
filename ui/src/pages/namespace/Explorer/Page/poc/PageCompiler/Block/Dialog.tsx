import * as DialogPrimitive from "@radix-ui/react-dialog";

import { Block, BlockPathType } from ".";

import { BlockList } from "./utils/BlockList";
import { Button } from "./Button";
import { DialogTrigger } from "~/design/Dialog";
import { DialogType } from "../../schema/blocks/dialog";
import { XIcon } from "lucide-react";
import { twMergeClsx } from "~/util/helpers";
import { useLocalDialogContainer } from "~/components/LocalDialog";
import { usePageEditor } from "../context/pageCompilerContext";
import { usePageEditorPanel } from "../../BlockEditor/EditorPanelProvider";

type DialogProps = {
  blockProps: DialogType;
  blockPath: BlockPathType;
  onOpenChange?: (open: boolean) => void;
};

const DialogBaseComponent = ({
  blockProps,
  blockPath,
  onOpenChange,
}: DialogProps) => {
  const { container } = useLocalDialogContainer();
  const { blocks, trigger } = blockProps;

  return (
    <DialogPrimitive.Root modal={false} onOpenChange={onOpenChange}>
      <DialogTrigger>
        <Button data-local-dialog-trigger blockProps={trigger} />
      </DialogTrigger>

      <DialogPrimitive.DialogPortal container={container}>
        <div
          className="absolute inset-0 flex items-start justify-center px-5 pt-16"
          onClick={(event) => event.stopPropagation()}
        >
          <div
            className="absolute inset-0 -m-2 bg-black/10 backdrop-blur-sm"
            onClick={(event) => event.stopPropagation()}
          />
          <DialogPrimitive.Content
            className={twMergeClsx(
              "pointer-events-auto fixed z-50 grid w-full gap-4 rounded-b-lg bg-gray-1 p-6 animate-in data-[state=open]:fade-in-90 data-[state=open]:slide-in-from-bottom-10 sm:max-w-lg sm:rounded-lg sm:zoom-in-90 data-[state=open]:sm:slide-in-from-bottom-0",
              "dark:bg-gray-dark-1"
            )}
            onInteractOutside={(e) => {
              e.preventDefault();
            }}
          >
            <DialogPrimitive.Close
              data-slot="dialog-close"
              className="absolute right-4 top-4 rounded-sm opacity-70 transition-opacity hover:opacity-100 focus:outline focus:outline-2 disabled:pointer-events-none [&_svg:not([class*='size-'])]:size-4 [&_svg]:pointer-events-none [&_svg]:shrink-0"
            >
              <XIcon />
            </DialogPrimitive.Close>
            <BlockList path={blockPath}>
              {blocks.map((block, index) => (
                <Block
                  key={index}
                  block={block}
                  blockPath={[...blockPath, index]}
                />
              ))}
            </BlockList>
          </DialogPrimitive.Content>
        </div>
      </DialogPrimitive.DialogPortal>
    </DialogPrimitive.Root>
  );
};

const EditModeDialog = (props: DialogProps) => {
  const { setPanel } = usePageEditorPanel();

  return (
    <DialogBaseComponent
      {...props}
      onOpenChange={() =>
        setPanel({
          action: "edit",
          path: props.blockPath,
          block: props.blockProps,
        })
      }
    />
  );
};

export const Dialog = (props: DialogProps) => {
  const { mode } = usePageEditor();

  if (mode === "edit") return <EditModeDialog {...props} />;

  return <DialogBaseComponent {...props} />;
};
