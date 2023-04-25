import * as React from "react";
import * as SelectPrimitive from "@radix-ui/react-select";

import Button, { ButtonProps } from "../Button";
import { RxCheck, RxChevronDown } from "react-icons/rx";

import clsx from "clsx";

// this component is mostly copied from https://ui.shadcn.com/docs/primitives/select

const Select = SelectPrimitive.Root;

const SelectGroup = SelectPrimitive.Group;

const SelectValue = SelectPrimitive.Value;

const SelectTrigger = React.forwardRef<
  React.ElementRef<typeof SelectPrimitive.Trigger>,
  React.ComponentPropsWithoutRef<typeof SelectPrimitive.Trigger> & ButtonProps
>(
  (
    { className, variant, size, children, disabled, block, loading, ...props },
    ref
  ) => (
    <SelectPrimitive.Trigger
      ref={ref}
      // border-opacity-20 will be deprecated in tailwind 3 but daisyui is still on tailwind 2
      // eslint-disable-next-line tailwindcss/migration-from-tailwind-2
      className={clsx(className)}
      {...props}
    >
      <Button
        variant={variant}
        size={size}
        disabled={disabled}
        block={block}
        loading={loading}
      >
        {children} <RxChevronDown />
      </Button>
    </SelectPrimitive.Trigger>
  )
);
SelectTrigger.displayName = SelectPrimitive.Trigger.displayName;

const SelectContent = React.forwardRef<
  React.ElementRef<typeof SelectPrimitive.Content>,
  React.ComponentPropsWithoutRef<typeof SelectPrimitive.Content>
>(({ className, children, ...props }, ref) => (
  <SelectPrimitive.Portal>
    <SelectPrimitive.Content
      ref={ref}
      className={clsx(
        "z-50 shadow-md",
        "rounded-md ring-1",
        "ring-black/5",
        "dark:ring-white/5",
        "border-gray-3 bg-gray-1",
        "dark:border-gray-dark-3 dark:bg-gray-dark-1",
        className
      )}
      {...props}
    >
      <SelectPrimitive.Viewport className="p-1">
        {children}
      </SelectPrimitive.Viewport>
    </SelectPrimitive.Content>
  </SelectPrimitive.Portal>
));

SelectContent.displayName = SelectPrimitive.Content.displayName;

const SelectLabel = React.forwardRef<
  React.ElementRef<typeof SelectPrimitive.Label>,
  React.ComponentPropsWithoutRef<typeof SelectPrimitive.Label>
>(({ className, ...props }, ref) => (
  <SelectPrimitive.Label
    ref={ref}
    className={clsx(
      "py-1.5 pr-2 pl-8 text-sm font-medium",
      "text-gray-8 ",
      "dark:text-gray-dark-8",
      className
    )}
    {...props}
  />
));
SelectLabel.displayName = SelectPrimitive.Label.displayName;

const SelectItem = React.forwardRef<
  React.ElementRef<typeof SelectPrimitive.Item>,
  React.ComponentPropsWithoutRef<typeof SelectPrimitive.Item>
>(({ className, children, ...props }, ref) => (
  <SelectPrimitive.Item
    ref={ref}
    className={clsx(
      "outline-nonedata-[disabled]:pointer-events-none relative flex cursor-default select-none items-center rounded-sm py-1.5 pr-2 pl-8 text-sm font-medium data-[disabled]:opacity-50",
      " focus:bg-gray-3 ",
      " dark:focus:bg-gray-dark-3 ",
      className
    )}
    {...props}
  >
    <span className="absolute left-2 flex h-3.5 w-3.5 items-center justify-center">
      <SelectPrimitive.ItemIndicator>
        <RxCheck className="h-4 w-4" />
      </SelectPrimitive.ItemIndicator>
    </span>

    <SelectPrimitive.ItemText>{children}</SelectPrimitive.ItemText>
  </SelectPrimitive.Item>
));
SelectItem.displayName = SelectPrimitive.Item.displayName;

const SelectSeparator = React.forwardRef<
  React.ElementRef<typeof SelectPrimitive.Separator>,
  React.ComponentPropsWithoutRef<typeof SelectPrimitive.Separator>
>(({ className, ...props }, ref) => (
  <SelectPrimitive.Separator
    ref={ref}
    className={clsx(
      "-mx-1 my-1 h-px",
      " bg-gray-3",
      " dark:bg-gray-dark-3",
      className
    )}
    {...props}
  />
));
SelectSeparator.displayName = SelectPrimitive.Separator.displayName;

export {
  Select,
  SelectGroup,
  SelectValue,
  SelectTrigger,
  SelectContent,
  SelectLabel,
  SelectItem,
  SelectSeparator,
};
