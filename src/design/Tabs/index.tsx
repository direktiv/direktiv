import * as React from "react";
import * as TabsPrimitive from "@radix-ui/react-tabs";

import clsx from "clsx";

const Tabs = TabsPrimitive.Root;

const TabsList = React.forwardRef<
  React.ElementRef<typeof TabsPrimitive.List>,
  React.ComponentPropsWithoutRef<typeof TabsPrimitive.List> & {
    varient?: "default" | "primary";
  }
>(({ className, varient = "default", ...props }, ref) => (
  <TabsPrimitive.List
    ref={ref}
    className={clsx(
      varient === "default" &&
        "inline-flex items-center justify-center rounded-md bg-slate-100 p-1 dark:bg-slate-800",
      varient === "primary" &&
        "inline-flex items-center justify-center rounded-md p-1",
      className
    )}
    {...props}
  />
));
TabsList.displayName = TabsPrimitive.List.displayName;

const TabsTrigger = React.forwardRef<
  React.ElementRef<typeof TabsPrimitive.Trigger>,
  React.ComponentPropsWithoutRef<typeof TabsPrimitive.Trigger> & {
    varient?: "default" | "primary";
  }
>(({ className, varient = "default", ...props }, ref) => (
  <TabsPrimitive.Trigger
    className={clsx(
      varient === "default" &&
        "inline-flex min-w-[100px] items-center justify-center rounded-[0.185rem] px-3 py-1.5  text-sm font-medium transition-all  disabled:pointer-events-none disabled:opacity-50  data-[state=active]:shadow-sm",
      varient === "default" &&
        "text-gray-10 data-[state=active]:bg-white data-[state=active]:text-gray-12 ",
      varient === "default" &&
        "dark:text-gray-dark-10 dark:data-[state=active]:bg-black dark:data-[state=active]:text-gray-dark-12",
      varient === "primary" &&
        "mx-4 flex items-center gap-x-2 whitespace-nowrap border-b-2 border-transparent px-1 pb-4 text-sm font-medium",
      varient === "primary" &&
        "text-gray-11 hover:border-gray-8 hover:text-gray-12",
      varient === "primary" &&
        "dark:text-gray-dark-11 dark:hover:border-gray-dark-8 dark:hover:text-gray-dark-12",
      varient === "primary" && "data-[state=active]:border-primary-500",
      className
    )}
    {...props}
    ref={ref}
  />
));
TabsTrigger.displayName = TabsPrimitive.Trigger.displayName;

const TabsContent = React.forwardRef<
  React.ElementRef<typeof TabsPrimitive.Content>,
  React.ComponentPropsWithoutRef<typeof TabsPrimitive.Content> & {
    noBorder?: boolean;
  }
>(({ className, noBorder = false, ...props }, ref) => (
  <TabsPrimitive.Content
    className={clsx(
      "mt-2 rounded-md p-6",
      !noBorder && "border border-gray-4  dark:border-gray-dark-4",
      className
    )}
    {...props}
    ref={ref}
  />
));
TabsContent.displayName = TabsPrimitive.Content.displayName;

export { Tabs, TabsList, TabsTrigger, TabsContent };
