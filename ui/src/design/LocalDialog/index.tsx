import * as DialogPrimitive from "@radix-ui/react-dialog";

import { PropsWithChildren, useCallback, useRef } from "react";

import { DialogProps } from "@radix-ui/react-dialog";
import { twMergeClsx } from "~/util/helpers";
import { useLocalDialogContainer } from "./container";

export const LocalDialog = ({
  children,
  open,
  onOpenChange,
}: PropsWithChildren & Pick<DialogProps, "onOpenChange" | "open">) => (
  <DialogPrimitive.Root modal={false} onOpenChange={onOpenChange} open={open}>
    {children}
  </DialogPrimitive.Root>
);

export const LocalDialogContent = ({ children }: PropsWithChildren) => {
  const { container } = useLocalDialogContainer();
  const rectRef = useRef<DOMRect | undefined>(undefined);

  const setRectCallback = useCallback((el: HTMLDivElement | null) => {
    rectRef.current = el?.getBoundingClientRect();
  }, []);

  return (
    <DialogPrimitive.DialogPortal container={container}>
      <div
        ref={setRectCallback}
        className="absolute inset-0 flex items-center justify-center px-5"
        onClick={(event) => event.stopPropagation()}
      >
        <div
          className="fixed inset-0 bg-black/10 backdrop-blur-sm"
          style={{
            width: rectRef.current?.width,
            height: rectRef.current?.height,
            top: rectRef.current?.top,
            left: rectRef.current?.left,
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
