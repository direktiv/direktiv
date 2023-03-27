"use client";

import * as CheckboxPrimitive from "@radix-ui/react-checkbox";
import * as React from "react";

import { Check } from "lucide-react";
import clsx from "clsx";

interface CustomCheckboxProps {
  disabled?: boolean;
  size?: "lg" | "md" | "sm" | "xs";
}
const Checkbox = React.forwardRef<
  React.ElementRef<typeof CheckboxPrimitive.Root>,
  React.ComponentPropsWithoutRef<typeof CheckboxPrimitive.Root> &
    CustomCheckboxProps
>(({ className, ...props }, ref) => (
  <CheckboxPrimitive.Root
    ref={ref}
    className={clsx(
      "peer shrink-0 rounded-sm border border-gray-4 focus:outline-none focus:ring-2 focus:ring-gray-7 focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 dark:border-gray-10 dark:text-gray-1 dark:focus:ring-gray-7 dark:focus:ring-offset-gray-12",
      props.size === "lg" && "h-5 w-5",
      props.size === "md" && "h-4 w-4",
      props.size === "sm" && "h-3 w-3",
      props.size === "xs" && "h-2 w-2",
      className
    )}
    {...props}
  >
    <CheckboxPrimitive.Indicator
      className={clsx("flex items-center justify-center")}
    >
      <Check
        className={clsx(
          props.size === "lg" && "h-5 w-5",
          (props.size === "md" || props.size === undefined) && "h-4 w-4",
          props.size === "sm" && "h-3 w-3",
          props.size === "xs" && "h-2 w-2"
        )}
      />
    </CheckboxPrimitive.Indicator>
  </CheckboxPrimitive.Root>
));
Checkbox.displayName = CheckboxPrimitive.Root.displayName;

export { Checkbox };
