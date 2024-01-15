import * as CheckboxPrimitive from "@radix-ui/react-checkbox";
import * as React from "react";

import { Check } from "lucide-react";
import { twMergeClsx } from "~/util/helpers";

interface CustomCheckboxProps {
  disabled?: boolean;
  size?: "sm" | "lg";
}
const Checkbox = React.forwardRef<
  React.ElementRef<typeof CheckboxPrimitive.Root>,
  React.ComponentPropsWithoutRef<typeof CheckboxPrimitive.Root> &
    CustomCheckboxProps
>(({ className, size, ...props }, ref) => (
  <CheckboxPrimitive.Root
    ref={ref}
    className={twMergeClsx(
      "peer shrink-0 rounded-sm border focus:outline-none focus:ring-2 focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50",
      "border-gray-11 text-gray-12 focus:ring-gray-7 focus:ring-offset-gray-1",
      "dark:border-gray-dark-11 dark:bg-black dark:text-gray-dark-12 dark:focus:ring-gray-dark-7 dark:focus:ring-offset-gray-dark-1",
      size === "lg" && "h-5 w-5",
      size === "sm" && "h-3 w-3",
      !size && "h-4 w-4",
      className
    )}
    {...props}
  >
    <CheckboxPrimitive.Indicator
      className={twMergeClsx("flex items-center justify-center")}
    >
      <Check
        className={twMergeClsx(
          size === "lg" && "h-5 w-5",
          size === "sm" && "h-3 w-3",
          !size && "h-4 w-4"
        )}
      />
    </CheckboxPrimitive.Indicator>
  </CheckboxPrimitive.Root>
));
Checkbox.displayName = CheckboxPrimitive.Root.displayName;

export { Checkbox };
