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
  variant?: "alwaysVisibleRight" | "alwaysVisibleLeft";
};

const HoverElement = React.forwardRef<
  HTMLDivElement,
  React.HTMLAttributes<HTMLDivElement> & HoverElementProps
>(({ className, children, variant }, ref) => (
  <div
    ref={ref}
    className={twMergeClsx(
      "absolute top-2 flex justify-end gap-1",
      !variant && [
        "invisible right-2 transition-opacity group-hover:visible group-focus:invisible",
      ],
      variant === "alwaysVisibleRight" && [
        "right-2 opacity-50 hover:opacity-100",
      ],
      variant === "alwaysVisibleLeft" && [
        "right-12 opacity-50 hover:opacity-100",
      ],
      className
    )}
  >
    {children}
  </div>
));
HoverElement.displayName = "HoverElement";

export { HoverContainer, HoverElement };
