"use client";

import * as CheckboxPrimitive from "@radix-ui/react-checkbox";
import * as React from "react";

import { Check } from "lucide-react";
import clsx from "clsx";

interface CustomCheckboxProps {
  disabled?: boolean;
  variant?:
    | "primary"
    | "secondary"
    | "accent"
    | "success"
    | "warning"
    | "info"
    | "error";
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
      "peer shrink-0 rounded-sm border border-gray-gray4 focus:outline-none focus:ring-2 focus:ring-gray-gray7 focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 dark:border-gray-gray10 dark:text-gray-gray1 dark:focus:ring-gray-gray7 dark:focus:ring-offset-gray-gray12",
      props.size === "lg" && "h-5 w-5",
      (props.size === "md" || props.size === undefined) && "h-4 w-4",
      props.size === "sm" && "h-3 w-3",
      props.size === "xs" && "h-2 w-2",
      props.variant === "primary" && "bg-primary-500",
      props.variant === "primary" && "bg-secondary",
      props.variant === "accent" && "bg-accent",
      props.variant === "warning" && "bg-warning",
      props.variant === "success" && "bg-success",
      props.variant === "error" && "bg-error",
      props.variant === "info" && "bg-info",
      className
    )}
    {...props}
  >
    <CheckboxPrimitive.Indicator
      className={clsx("flex items-center justify-center")}
    >
      <Check
        className={clsx(
          "stroke-white",
          props.size === "lg" && "h-5 w-5",
          (props.size === "md" || props.size === undefined) && "h-4 w-4",
          props.size === "sm" && "h-3 w-3",
          props.size === "xs" && "h-2 w-2",
          (props.variant === "secondary" || props.variant === undefined) &&
            "stroke-black-alpha-700"
        )}
      />
    </CheckboxPrimitive.Indicator>
  </CheckboxPrimitive.Root>
));
Checkbox.displayName = CheckboxPrimitive.Root.displayName;

export { Checkbox };
