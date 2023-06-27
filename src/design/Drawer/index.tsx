import * as DrawerPrimitive from "@radix-ui/react-dialog";
import * as React from "react";

import { twMergeClsx } from "~/util/helpers";

const Drawer = DrawerPrimitive.Root;

const DrawerTrigger = DrawerPrimitive.Trigger;

interface DrawerPortalProps extends DrawerPrimitive.DialogPortalProps {
  position: "top" | "bottom" | "left" | "right";
}

const DrawerPortal = ({
  position = "right",
  className,
  children,
  ...props
}: DrawerPortalProps) => (
  <DrawerPrimitive.Portal className={twMergeClsx(className)} {...props}>
    <div
      className={twMergeClsx(
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
    className={twMergeClsx(
      "fixed inset-0 z-50 bg-black-alpha-9 backdrop-grayscale-0 transition-all duration-100 data-[state=closed]:animate-out data-[state=closed]:fade-out data-[state=open]:fade-in dark:bg-white-alpha-9",
      className
    )}
    {...props}
    ref={ref}
  />
));
DrawerOverlay.displayName = DrawerPrimitive.Overlay.displayName;

const DrawerContent = React.forwardRef<
  React.ElementRef<typeof DrawerPrimitive.Content>,
  React.ComponentPropsWithoutRef<typeof DrawerPrimitive.Content>
>(({ className, children, ...props }, ref) => (
  <DrawerPortal position="left">
    <DrawerOverlay />
    <DrawerPrimitive.Content
      ref={ref}
      className={twMergeClsx(
        "fixed inset-0 z-50 scale-100 gap-4 bg-gray-1 opacity-100 shadow-lg dark:bg-gray-dark-1 ",
        "h-full animate-in slide-in-from-left duration-300",
        "w-52 p-4 focus:outline-none",
        className
      )}
      {...props}
    >
      {children}
    </DrawerPrimitive.Content>
  </DrawerPortal>
));
DrawerContent.displayName = DrawerPrimitive.Content.displayName;

const DrawerHeader = ({
  className,
  ...props
}: React.HTMLAttributes<HTMLDivElement>) => (
  <div
    className={twMergeClsx(
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
    className={twMergeClsx(
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
    className={twMergeClsx(
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
    className={twMergeClsx("text-sm text-black dark:text-white", className)}
    {...props}
  />
));
DrawerDescription.displayName = DrawerPrimitive.Description.displayName;

export {
  Drawer,
  DrawerTrigger,
  DrawerContent,
  DrawerHeader,
  DrawerFooter,
  DrawerTitle,
  DrawerDescription,
};
