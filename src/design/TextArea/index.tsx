import * as React from "react";

import clsx from "clsx";

export interface TextareaProps
  extends React.TextareaHTMLAttributes<HTMLTextAreaElement> {
  custom?: boolean; // just to avoid  error  An interface declaring no members is equivalent to its supertype  @typescript-eslint/no-empty-interface
}

const Textarea = React.forwardRef<HTMLTextAreaElement, TextareaProps>(
  ({ className, ...props }, ref) => (
    <textarea
      className={clsx(
        "flex h-20 w-full rounded-md border bg-transparent py-2 px-3 text-sm focus:outline-none focus:ring-2 focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50",
        "border-gray-4 text-gray-12 placeholder:text-gray-7 focus:ring-gray-7 focus:ring-offset-gray-1",
        "dark:border-gray-dark-4 dark:text-gray-dark-12 dark:focus:ring-gray-dark-7 dark:focus:ring-offset-gray-dark-1",
        className
      )}
      ref={ref}
      {...props}
    />
  )
);
Textarea.displayName = "Textarea";

export { Textarea };
