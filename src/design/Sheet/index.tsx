import * as React from "react";
import * as SheetPrimitive from "@radix-ui/react-dialog";

import { X } from "lucide-react";
import clsx from "clsx";

const Sheet = SheetPrimitive.Root;

const SheetTrigger = SheetPrimitive.Trigger;

/* eslint-disable-next-line */
interface SheetPortalProps extends SheetPrimitive.DialogPortalProps {
  position: "top" | "bottom" | "left" | "right";
}

const SheetPortal = ({
  position = "right",
  className,
  children,
  ...props
}: SheetPortalProps) => (
  <SheetPrimitive.Portal className={clsx(className)} {...props}>
    <div
      className={clsx(
        position === "top" && "items-start",
        position === "bottom" && "items-end",
        position === "left" && "justify-start",
        position === "right" && "justify-end"
      )}
    >
      {children}
    </div>
  </SheetPrimitive.Portal>
);
SheetPortal.displayName = SheetPrimitive.Portal.displayName;

const SheetOverlay = React.forwardRef<
  React.ElementRef<typeof SheetPrimitive.Overlay>,
  React.ComponentPropsWithoutRef<typeof SheetPrimitive.Overlay>
>(({ className, ...props }, ref) => (
  <SheetPrimitive.Overlay
    className={clsx(
      "fixed inset-0 z-50 bg-black/30 backdrop-grayscale-0 transition-all duration-100 data-[state=closed]:animate-out data-[state=closed]:fade-out data-[state=open]:fade-in",
      className
    )}
    {...props}
    ref={ref}
  />
));
SheetOverlay.displayName = SheetPrimitive.Overlay.displayName;

const getSizeClass = (position: string, size: string): string => {
  let sizeClass = "";
  if (["top", "bottom"].indexOf(position) > -1) {
    switch (size) {
      case "content":
        sizeClass = "max-h-screen";
        break;
      case "default":
        sizeClass = "h-1/3";
        break;
      case "sm":
        sizeClass = "h-1/4";
        break;
      case "lg":
        sizeClass = "h-1/2";
        break;
      case "xl":
        sizeClass = "h-5/6";
        break;
      case "full":
        sizeClass = "h-screen";
        break;
    }
  } else if (["left", "right"].indexOf(position) > -1) {
    switch (size) {
      case "content":
        sizeClass = "max-w-screen";
        break;
      case "default":
        sizeClass = "w-1/3";
        break;
      case "sm":
        sizeClass = "w-1/4";
        break;
      case "lg":
        sizeClass = "w-1/2";
        break;
      case "xl":
        sizeClass = "w-5/6";
        break;
      case "full":
        sizeClass = "w-screen";
        break;
    }
  }
  return sizeClass;
};
export interface DialogContentProps
  extends React.ComponentPropsWithoutRef<typeof SheetPrimitive.Content> {
  position?: "top" | "bottom" | "left" | "right";
  size?: "default" | "content" | "sm" | "lg" | "xl" | "full";
  noClose?: boolean;
}

const SheetContent = React.forwardRef<
  React.ElementRef<typeof SheetPrimitive.Content>,
  DialogContentProps
>(
  (
    {
      position = "left",
      size = "default",
      className,
      children,
      noClose,
      ...props
    },
    ref
  ) => (
    <SheetPortal position={position}>
      <SheetOverlay />
      <SheetPrimitive.Content
        ref={ref}
        className={clsx(
          "fixed inset-0 z-50 scale-100 gap-4 bg-white p-4 opacity-100 shadow-lg dark:bg-black ",
          position === "top" &&
            "w-full animate-in slide-in-from-top duration-300",
          position === "bottom" &&
            "w-full animate-in slide-in-from-bottom duration-300",
          position === "left" &&
            "h-full animate-in slide-in-from-left duration-300",
          position === "right" &&
            "h-full animate-in slide-in-from-right duration-300",
          getSizeClass(position, size),
          className
        )}
        {...props}
      >
        {children}
        {!noClose && (
          <SheetPrimitive.Close className="ring-offset-background focus:ring-ring data-[state=open]:bg-secondary absolute right-4 top-4 rounded-sm opacity-70 transition-opacity hover:opacity-100 focus:outline-none focus:ring-2 focus:ring-offset-2 disabled:pointer-events-none">
            <X className="h-4 w-4" />
            <span className="sr-only">Close</span>
          </SheetPrimitive.Close>
        )}
      </SheetPrimitive.Content>
    </SheetPortal>
  )
);
SheetContent.displayName = SheetPrimitive.Content.displayName;

const SheetHeader = ({
  className,
  ...props
}: React.HTMLAttributes<HTMLDivElement>) => (
  <div
    className={clsx(
      "flex flex-col space-y-2 text-center sm:text-left",
      className
    )}
    {...props}
  />
);
SheetHeader.displayName = "SheetHeader";

const SheetFooter = ({
  className,
  ...props
}: React.HTMLAttributes<HTMLDivElement>) => (
  <div
    className={clsx(
      "flex flex-col-reverse sm:flex-row sm:justify-end sm:space-x-2",
      className
    )}
    {...props}
  />
);
SheetFooter.displayName = "SheetFooter";

const SheetTitle = React.forwardRef<
  React.ElementRef<typeof SheetPrimitive.Title>,
  React.ComponentPropsWithoutRef<typeof SheetPrimitive.Title>
>(({ className, ...props }, ref) => (
  <SheetPrimitive.Title
    ref={ref}
    className={clsx(
      "text-lg font-semibold text-black dark:text-white",
      className
    )}
    {...props}
  />
));
SheetTitle.displayName = SheetPrimitive.Title.displayName;

const SheetDescription = React.forwardRef<
  React.ElementRef<typeof SheetPrimitive.Description>,
  React.ComponentPropsWithoutRef<typeof SheetPrimitive.Description>
>(({ className, ...props }, ref) => (
  <SheetPrimitive.Description
    ref={ref}
    className={clsx("text-sm text-black dark:text-white", className)}
    {...props}
  />
));
SheetDescription.displayName = SheetPrimitive.Description.displayName;

const SheetMain = React.forwardRef<
  HTMLDivElement,
  React.HTMLAttributes<HTMLDivElement>
>(({ className, ...props }, ref) => (
  <div
    ref={ref}
    className={clsx("h-full bg-white dark:bg-black", className)}
    {...props}
  />
));
SheetMain.displayName = "SheetMain";

export {
  Sheet,
  SheetMain,
  SheetTrigger,
  SheetContent,
  SheetHeader,
  SheetFooter,
  SheetTitle,
  SheetDescription,
};
