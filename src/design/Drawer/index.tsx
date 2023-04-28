import * as DrawerPrimitive from "@radix-ui/react-dialog";
import * as React from "react";

import { X } from "lucide-react";
import clsx from "clsx";

const Drawer = DrawerPrimitive.Root;

const DrawerTrigger = DrawerPrimitive.Trigger;

/* eslint-disable-next-line */
interface DrawerPortalProps extends DrawerPrimitive.DialogPortalProps {
  position: "top" | "bottom" | "left" | "right";
}

const DrawerPortal = ({
  position = "right",
  className,
  children,
  ...props
}: DrawerPortalProps) => (
  <DrawerPrimitive.Portal className={clsx(className)} {...props}>
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
  </DrawerPrimitive.Portal>
);
DrawerPortal.displayName = DrawerPrimitive.Portal.displayName;

const DrawerOverlay = React.forwardRef<
  React.ElementRef<typeof DrawerPrimitive.Overlay>,
  React.ComponentPropsWithoutRef<typeof DrawerPrimitive.Overlay>
>(({ className, ...props }, ref) => (
  <DrawerPrimitive.Overlay
    className={clsx(
      "fixed inset-0 z-50 bg-black/30 backdrop-grayscale-0 transition-all duration-100 data-[state=closed]:animate-out data-[state=closed]:fade-out data-[state=open]:fade-in",
      className
    )}
    {...props}
    ref={ref}
  />
));
DrawerOverlay.displayName = DrawerPrimitive.Overlay.displayName;

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
  extends React.ComponentPropsWithoutRef<typeof DrawerPrimitive.Content> {
  position?: "top" | "bottom" | "left" | "right";
  size?: "default" | "content" | "sm" | "lg" | "xl" | "full";
  noClose?: boolean;
}

const DrawerContent = React.forwardRef<
  React.ElementRef<typeof DrawerPrimitive.Content>,
  DialogContentProps
>(
  (
    {
      position = "left",
      size = "sm",
      className,
      children,
      noClose = true,
      ...props
    },
    ref
  ) => (
    <DrawerPortal position={position}>
      <DrawerOverlay />
      <DrawerPrimitive.Content
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
          <DrawerPrimitive.Close className="ring-offset-background focus:ring-ring data-[state=open]:bg-secondary absolute right-4 top-4 rounded-sm opacity-70 transition-opacity hover:opacity-100 focus:outline-none focus:ring-2 focus:ring-offset-2 disabled:pointer-events-none">
            <X className="h-4 w-4" />
            <span className="sr-only">Close</span>
          </DrawerPrimitive.Close>
        )}
      </DrawerPrimitive.Content>
    </DrawerPortal>
  )
);
DrawerContent.displayName = DrawerPrimitive.Content.displayName;

const DrawerHeader = ({
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
DrawerHeader.displayName = "DrawerHeader";

const DrawerFooter = ({
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
DrawerFooter.displayName = "DrawerFooter";

const DrawerTitle = React.forwardRef<
  React.ElementRef<typeof DrawerPrimitive.Title>,
  React.ComponentPropsWithoutRef<typeof DrawerPrimitive.Title>
>(({ className, ...props }, ref) => (
  <DrawerPrimitive.Title
    ref={ref}
    className={clsx(
      "text-lg font-semibold text-black dark:text-white",
      className
    )}
    {...props}
  />
));
DrawerTitle.displayName = DrawerPrimitive.Title.displayName;

const DrawerDescription = React.forwardRef<
  React.ElementRef<typeof DrawerPrimitive.Description>,
  React.ComponentPropsWithoutRef<typeof DrawerPrimitive.Description>
>(({ className, ...props }, ref) => (
  <DrawerPrimitive.Description
    ref={ref}
    className={clsx("text-sm text-black dark:text-white", className)}
    {...props}
  />
));
DrawerDescription.displayName = DrawerPrimitive.Description.displayName;

const DrawerMain = React.forwardRef<
  HTMLDivElement,
  React.HTMLAttributes<HTMLDivElement>
>(({ className, ...props }, ref) => (
  <div
    ref={ref}
    className={clsx("h-full bg-white dark:bg-black", className)}
    {...props}
  />
));
DrawerMain.displayName = "DrawerMain";

export {
  Drawer,
  DrawerMain,
  DrawerTrigger,
  DrawerContent,
  DrawerHeader,
  DrawerFooter,
  DrawerTitle,
  DrawerDescription,
};
