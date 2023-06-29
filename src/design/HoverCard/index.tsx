import * as HoverCardPrimitive from "@radix-ui/react-hover-card";
import * as React from "react";

import { twMergeClsx } from "~/util/helpers";

const HoverCard = HoverCardPrimitive.Root;
const HoverCardTrigger = HoverCardPrimitive.Trigger;

type HoverCardProps = typeof HoverCardPrimitive.Content;
type AdditionalHoverCardContentProps = { noBackground?: boolean };

const HoverCardContent = React.forwardRef<
  React.ElementRef<HoverCardProps>,
  React.ComponentPropsWithoutRef<HoverCardProps> &
    AdditionalHoverCardContentProps
>(
  (
    {
      className,
      align = "center",
      sideOffset = 4,
      noBackground = false,
      ...props
    },
    ref
  ) => (
    <HoverCardPrimitive.Content
      ref={ref}
      align={align}
      sideOffset={sideOffset}
      className={twMergeClsx(
        "z-50 rounded-md border p-4 shadow-md outline-none animate-in zoom-in-90",
        "border-gray-3",
        "dark:border-gray-dark-3",
        !noBackground && "bg-white dark:bg-black",
        className
      )}
      {...props}
    />
  )
);
HoverCardContent.displayName = HoverCardPrimitive.Content.displayName;

export { HoverCard, HoverCardTrigger, HoverCardContent };
