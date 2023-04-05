"use client";

import * as React from "react";
import * as SliderPrimitive from "@radix-ui/react-slider";

import clsx from "clsx";

const Slider = React.forwardRef<
  React.ElementRef<typeof SliderPrimitive.Root>,
  React.ComponentPropsWithoutRef<typeof SliderPrimitive.Root>
>(({ className, ...props }, ref) => (
  <SliderPrimitive.Root
    ref={ref}
    className={clsx(
      "relative flex w-full touch-none select-none items-center",
      className
    )}
    {...props}
  >
    <SliderPrimitive.Track
      className={clsx(
        "relative h-2 w-full grow overflow-hidden rounded-full",
        "bg-gray-3",
        "dark:bg-gray-dark-3"
      )}
    >
      <SliderPrimitive.Range
        className={clsx("absolute h-full", "bg-gray-12", "dark:bg-gray-dark-9")}
      />
    </SliderPrimitive.Track>
    <SliderPrimitive.Thumb
      className={clsx(
        "block h-5 w-5 rounded-full ",
        "transition-colors focus:outline-none",
        "border-2 focus:ring-2 focus:ring-offset-2 disabled:pointer-events-none  disabled:opacity-50",
        "border-gray-12 bg-gray-1 focus:ring-gray-7",
        "dark:border-gray-dark-12 dark:bg-slate-400 dark:focus:ring-gray-dark-7 dark:focus:ring-offset-gray-12"
      )}
    />
  </SliderPrimitive.Root>
));
Slider.displayName = SliderPrimitive.Root.displayName;

export { Slider };
