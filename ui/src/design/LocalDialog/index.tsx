// eslint-disable sort-imports
import * as DialogPrimitive from "@radix-ui/react-dialog";

import { PropsWithChildren } from "react";

import { twMergeClsx } from "~/util/helpers";
import { useLocalDialogContainer } from "./container";
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

  return (
    <DialogPrimitive.DialogPortal container={container}>
      <div
        className="absolute inset-0 flex items-center justify-center px-5"
        onClick={(event) => event.stopPropagation()}
      >
        <div
          className="fixed inset-0 bg-black/10 backdrop-blur-sm"
          style={{
            width: rect?.width,
            height: rect?.height,
            top: rect?.top,
            left: rect?.left,
          }}
        />
        <DialogPrimitive.Content
          className={twMergeClsx(
            "pointer-events-auto fixed z-50 grid w-full gap-4 rounded-b-lg bg-gray-1 p-6 animate-in data-[state=open]:fade-in-90 data-[state=open]:slide-in-from-bottom-10 sm:max-w-lg sm:rounded-lg sm:zoom-in-90 data-[state=open]:sm:slide-in-from-bottom-0",
            "dark:bg-gray-dark-1"
          )}
          onInteractOutside={(event) => {
            event.preventDefault();
          }}
        >
          {children}
        </DialogPrimitive.Content>
      </div>
    </DialogPrimitive.DialogPortal>
  );
};
