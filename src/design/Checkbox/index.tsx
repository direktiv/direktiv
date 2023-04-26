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
>(({ className, size = "md", ...props }, ref) => (
  <CheckboxPrimitive.Root
    ref={ref}
    className={clsx(
      "peer shrink-0 rounded-sm borderfocus:outline-none focus:ring-2  border focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 ",
      "text-gray-12 border-gray-11 focus:ring-gray-7 focus:ring-offset-gray-1 ",
      "dark:text-gray-dark-12 dark:border-gray-dark-11 dark:focus:ring-gray-dark-7 dark:focus:ring-offset-gray-dark-1 dark:bg-black ",
      size === "lg" && "h-5 w-5",
      size === "md" && "h-4 w-4",
      size === "sm" && "h-3 w-3",
      size === "xs" && "h-2 w-2",
      className
    )}
    {...props}
  >
    <CheckboxPrimitive.Indicator
      className={clsx("flex items-center justify-center")}
    >
      <Check
        className={clsx(
          size === "lg" && "h-5 w-5",
          size === "md" && "h-4 w-4",
          size === "sm" && "h-3 w-3",
          size === "xs" && "h-2 w-2"
        )}
      />
    </CheckboxPrimitive.Indicator>
  </CheckboxPrimitive.Root>
));
Checkbox.displayName = CheckboxPrimitive.Root.displayName;

export { Checkbox };
