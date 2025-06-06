import * as React from "react";
import * as SwitchPrimitives from "@radix-ui/react-switch";

import { twMergeClsx } from "~/util/helpers";

const Switch = React.forwardRef<
  React.ElementRef<typeof SwitchPrimitives.Root>,
  React.ComponentPropsWithoutRef<typeof SwitchPrimitives.Root>
>(({ className, ...props }, ref) => (
  <SwitchPrimitives.Root
    className={twMergeClsx(
      "peer inline-flex h-[24px] w-[44px] shrink-0 cursor-pointer items-center rounded-full border-2 border-transparent transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50",
      "focus:ring-gray-4 focus:ring-offset-gray-1 data-[state=checked]:bg-gray-12 data-[state=unchecked]:bg-gray-8",
      "dark:focus:ring-gray-dark-4 dark:focus:ring-offset-gray-dark-1 dark:data-[state=checked]:bg-gray-dark-12 dark:data-[state=unchecked]:bg-gray-dark-8",
      className
    )}
    {...props}
    ref={ref}
  >
    <SwitchPrimitives.Thumb
      className={twMergeClsx(
        "pointer-events-none block size-5 rounded-full shadow-lg ring-0 transition-transform data-[state=checked]:translate-x-5 data-[state=unchecked]:translate-x-0",
        "bg-gray-1",
        "dark:bg-gray-dark-1"
      )}
    />
  </SwitchPrimitives.Root>
));
Switch.displayName = SwitchPrimitives.Root.displayName;

export { Switch };
