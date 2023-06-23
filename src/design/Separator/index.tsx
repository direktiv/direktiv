import * as React from "react";
import * as SeparatorPrimitive from "@radix-ui/react-separator";

import { twMergeClsx } from "~/util/helpers";

type Props = typeof SeparatorPrimitive.Root;

const Separator = React.forwardRef<
  React.ElementRef<Props>,
  Omit<React.ComponentPropsWithoutRef<Props>, "orientation"> & {
    vertical?: boolean;
  }
>(({ className, vertical, ...props }, ref) => (
  <SeparatorPrimitive.Root
    ref={ref}
    orientation={vertical ? "vertical" : "horizontal"}
    className={twMergeClsx(
      "bg-gray-4",
      " dark:bg-gray-dark-4",
      vertical ? "h-full w-[1px]" : "h-[1px] w-full",
      className
    )}
    {...props}
  />
));
Separator.displayName = SeparatorPrimitive.Root.displayName;

export { Separator };
