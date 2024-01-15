import * as React from "react";
import * as SliderPrimitive from "@radix-ui/react-slider";

import { twMergeClsx } from "~/util/helpers";

const Slider = React.forwardRef<
  React.ElementRef<typeof SliderPrimitive.Root>,
  React.ComponentPropsWithoutRef<typeof SliderPrimitive.Root>
>(({ className, disabled, ...props }, ref) => (
  <SliderPrimitive.Root
    ref={ref}
    className={twMergeClsx(
      "relative flex w-full touch-none select-none items-center",
      disabled && "cursor-not-allowed opacity-40",
      className
    )}
    disabled={disabled}
    {...props}
  >
    <SliderPrimitive.Track
      className={twMergeClsx(
        "relative h-2 w-full grow overflow-hidden rounded-full",
        "bg-gray-3",
        "dark:bg-gray-dark-3"
      )}
    >
      <SliderPrimitive.Range
        className={twMergeClsx(
          "absolute h-full",
          "bg-gray-12",
          "dark:bg-gray-dark-12"
        )}
      />
    </SliderPrimitive.Track>
    <SliderPrimitive.Thumb
      className={twMergeClsx(
        "block h-5 w-5 rounded-full ",
        "transition-colors focus:outline-none",
        "border-2 focus:ring-2 focus:ring-offset-2",
        "border-gray-12 bg-gray-1 focus:ring-gray-7 dark:focus:ring-offset-gray-1",
        "dark:border-gray-dark-12 dark:bg-gray-dark-1 dark:focus:ring-gray-dark-7 dark:focus:ring-offset-gray-dark-1"
      )}
    />
  </SliderPrimitive.Root>
));
Slider.displayName = SliderPrimitive.Root.displayName;

export { Slider };
