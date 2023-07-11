import * as React from "react";
import * as TabsPrimitive from "@radix-ui/react-tabs";

import { twMergeClsx } from "~/util/helpers";

const Tabs = TabsPrimitive.Root;

const TabsList = React.forwardRef<
  React.ElementRef<typeof TabsPrimitive.List>,
  React.ComponentPropsWithoutRef<typeof TabsPrimitive.List> & {
    variant?: "boxed";
  }
>(({ className, variant, ...props }, ref) => (
  <TabsPrimitive.List
    ref={ref}
    className={twMergeClsx(
      variant === "boxed" &&
        "inline-flex items-start justify-center rounded-md bg-gray-2 p-1 dark:bg-gray-dark-2",
      !variant && "inline-flex items-center justify-center gap-x-8 rounded-md",
      className
    )}
    {...props}
  />
));
TabsList.displayName = TabsPrimitive.List.displayName;

const TabsTrigger = React.forwardRef<
  React.ElementRef<typeof TabsPrimitive.Trigger>,
  React.ComponentPropsWithoutRef<typeof TabsPrimitive.Trigger> & {
    variant?: "boxed";
  }
>(({ className, variant, ...props }, ref) => (
  <TabsPrimitive.Trigger
    className={twMergeClsx(
      variant === "boxed" &&
        "inline-flex min-w-[100px] items-center justify-center gap-x-2 rounded-[0.185rem] px-3  py-1.5 text-sm font-medium  transition-all disabled:pointer-events-none  disabled:opacity-50 data-[state=active]:shadow-sm",
      variant === "boxed" &&
        "text-gray-10 data-[state=active]:bg-white data-[state=active]:text-gray-12 ",
      variant === "boxed" &&
        "dark:text-gray-dark-10 dark:data-[state=active]:bg-black dark:data-[state=active]:text-gray-dark-12",
      !variant &&
        "flex items-center gap-x-2 whitespace-nowrap border-b-2 border-transparent px-1 pb-4 text-sm font-medium",
      !variant && "hover:border-current",
      !variant &&
        "data-[state=active]:border-primary-500 data-[state=active]:text-primary-500 [&>svg]:h-4 [&>svg]:w-auto",
      className
    )}
    {...props}
    ref={ref}
  />
));
TabsTrigger.displayName = TabsPrimitive.Trigger.displayName;

const TabsContent = React.forwardRef<
  React.ElementRef<typeof TabsPrimitive.Content>,
  React.ComponentPropsWithoutRef<typeof TabsPrimitive.Content>
>(({ className, ...props }, ref) => (
  <TabsPrimitive.Content
    className={twMergeClsx("mt-2", className)}
    {...props}
    ref={ref}
  />
));
TabsContent.displayName = TabsPrimitive.Content.displayName;

export { Tabs, TabsList, TabsTrigger, TabsContent };
