import * as HoverCardPrimitive from "@radix-ui/react-hover-card";
import * as React from "react";

import { twMergeClsx } from "~/util/helpers";

const HoverCard = HoverCardPrimitive.Root;

const HoverCardTrigger = HoverCardPrimitive.Trigger;

const HoverCardContent = React.forwardRef<
  React.ElementRef<typeof HoverCardPrimitive.Content>,
  React.ComponentPropsWithoutRef<typeof HoverCardPrimitive.Content>
>(({ className, align = "center", sideOffset = 4, ...props }, ref) => (
  <HoverCardPrimitive.Content
    ref={ref}
    align={align}
    sideOffset={sideOffset}
    className={twMergeClsx(
      "z-50 rounded-md border p-4 shadow-md outline-none animate-in zoom-in-90",
      "border-gray-3 bg-white",
      "dark:border-gray-dark-3 dark:bg-black",
      className
    )}
    {...props}
  />
));
HoverCardContent.displayName = HoverCardPrimitive.Content.displayName;

export { HoverCard, HoverCardTrigger, HoverCardContent };
