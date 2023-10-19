import * as ProgressPrimitive from "@radix-ui/react-progress";
import * as React from "react";

import { twMergeClsx } from "~/util/helpers";

const Progress = React.forwardRef<
  React.ElementRef<typeof ProgressPrimitive.Root>,
  React.ComponentPropsWithoutRef<typeof ProgressPrimitive.Root>
>(({ className, value, ...props }, ref) => (
  <ProgressPrimitive.Root
    ref={ref}
    className={twMergeClsx(
      "relative h-4 w-full overflow-hidden rounded-full bg-gray-2 dark:bg-gray-dark-2",
      className
    )}
    {...props}
  >
    <ProgressPrimitive.Indicator
      className="h-full w-full flex-1 bg-gray-12 transition-all dark:bg-gray-dark-12"
      style={{ transform: `translateX(-${100 - (value || 0)}%)` }}
    />
  </ProgressPrimitive.Root>
));
Progress.displayName = ProgressPrimitive.Root.displayName;

export { Progress };
