import React from "react";
import clsx from "clsx";

const Input = React.forwardRef<
  HTMLInputElement,
  React.InputHTMLAttributes<HTMLInputElement>
>(({ className, ...props }, ref) => (
  <input
    type="text"
    ref={ref}
    className={clsx(
      className,
      "flex h-9 w-full rounded-md border py-2 px-3 text-sm focus:outline-none focus:ring-2 focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50",
      "border-gray-4 bg-white text-black placeholder:text-gray-8 focus:ring-gray-4 focus:ring-offset-gray-1",
      "dark:border-gray-dark-4 dark:bg-black dark:text-white dark:placeholder:text-gray-dark-8 dark:focus:ring-gray-dark-4 dark:focus:ring-offset-gray-dark-1"
    )}
    {...props}
  />
));

Input.displayName = "Input";

export default Input;
