import React from "react";
import { twMergeClsx } from "~/util/helpers";

export const inputClasses = ({
  type,
  className,
}: {
  type?: React.HTMLInputTypeAttribute;
  className?: string;
}) =>
  twMergeClsx(
    "flex h-9 w-full rounded-md border bg-transparent px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50",
    "border-gray-4 placeholder:text-gray-8 focus:ring-gray-4 focus:ring-offset-gray-1",
    "dark:border-gray-dark-4 dark:placeholder:text-gray-dark-8 dark:focus:ring-gray-dark-4 dark:focus:ring-offset-gray-dark-1",
    type === "file" &&
      "p-0 file:mr-5 file:rounded-l-md file:border-0 file:px-6 file:py-2 file:text-sm file:font-medium",
    type === "file" &&
      "file:bg-black file:text-white dark:file:bg-white dark:file:text-black",
    className
  );

const Input = React.forwardRef<
  HTMLInputElement,
  React.InputHTMLAttributes<HTMLInputElement>
>(({ className, type, ...props }, ref) => (
  <input
    ref={ref}
    type={type}
    className={inputClasses({ type, className })}
    {...props}
  />
));

Input.displayName = "Input";

export default Input;
