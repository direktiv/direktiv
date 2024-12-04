import React from "react";
import { twMergeClsx } from "~/util/helpers";

const HoverContainer = React.forwardRef<
  HTMLDivElement,
  React.HTMLAttributes<HTMLDivElement>
>(({ className, children, ...props }, ref) => (
  <div
    {...props}
    ref={ref}
    className={twMergeClsx("group relative", className)}
  >
    {children}
  </div>
));
HoverContainer.displayName = "HoverContainer";

export type HoverElementProps = {
  variant?: "alwaysVisible";
};

const HoverElement = React.forwardRef<
  HTMLDivElement,
  React.HTMLAttributes<HTMLDivElement> & HoverElementProps
>(({ className, children, variant }, ref) => (
  <div
    ref={ref}
    className={twMergeClsx(
      "absolute right-2 top-2 flex justify-end gap-1",
      !variant && [
        "invisible transition-opacity group-hover:visible group-focus:invisible",
      ],
      variant === "alwaysVisible" && ["opacity-50 hover:opacity-100"],
      className
    )}
  >
    {children}
  </div>
));
HoverElement.displayName = "HoverElement";

export { HoverContainer, HoverElement };
