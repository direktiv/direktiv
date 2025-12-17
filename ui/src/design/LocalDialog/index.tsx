import * as DialogPrimitive from "@radix-ui/react-dialog";

import { PropsWithChildren } from "react";
import { twMergeClsx } from "~/util/helpers";
import { useLocalDialogContainer } from "./container";
import { usePageStateContext } from "~/pages/namespace/Explorer/Page/poc/PageCompiler/context/pageCompilerContext";
import { useParentRect } from "./useParentRect";

export const LocalDialog = ({
  children,
  open,
  onOpenChange,
}: PropsWithChildren &
  Pick<
    React.ComponentProps<typeof DialogPrimitive.Root>,
    "onOpenChange" | "open"
  >) => (
  <DialogPrimitive.Root modal={false} onOpenChange={onOpenChange} open={open}>
    {children}
  </DialogPrimitive.Root>
);

export const LocalDialogContent = ({ children }: PropsWithChildren) => {
  const { container } = useLocalDialogContainer();
  const rect = useParentRect(container);
  const { scrollPos } = usePageStateContext();

  return (
    <DialogPrimitive.DialogPortal container={container}>
      <div
        className="absolute inset-0 flex items-center justify-center"
        onClick={(event) => event.stopPropagation()}
      >
        <div
          className="fixed bg-black/10 backdrop-blur-sm max-lg:absolute max-lg:!inset-0"
          style={{
            width: rect?.width,
            height: rect?.height,
            top: rect?.top,
            left: rect?.left,
          }}
        />
        <DialogPrimitive.Content
          className={twMergeClsx(
            "pointer-events-auto absolute inset-x-0 top-12 z-50 grid gap-4 rounded-md rounded-b-lg bg-gray-1 p-6 animate-in",
            "data-[state=open]:fade-in-90 data-[state=open]:slide-in-from-bottom-10 dark:bg-gray-dark-1",
            "sm:inset-x-12 sm:mx-auto sm:max-w-lg sm:rounded-lg sm:zoom-in-90 data-[state=open]:sm:slide-in-from-bottom-0"
          )}
          onInteractOutside={(event) => {
            event.preventDefault();
          }}
          style={{
            top: scrollPos + 48,
          }}
        >
          {children}
        </DialogPrimitive.Content>
      </div>
    </DialogPrimitive.DialogPortal>
  );
};
