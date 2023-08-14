import React from "react";
import { twMergeClsx } from "~/util/helpers";

interface LogEntryProps extends React.HTMLAttributes<HTMLDivElement> {
  time?: string;
  variant?: "success" | "error" | "warning" | "info";
}
export const Logs = React.forwardRef<
  HTMLDivElement,
  React.HTMLAttributes<HTMLDivElement> & {
    wordWrap?: boolean;
  }
>(({ className, children, wordWrap, ...props }, ref) => (
  <div className="overflow-x-auto">
    <div
      ref={ref}
      {...props}
      className={twMergeClsx(
        !wordWrap && "w-max",
        !wordWrap && "[&>div>div>pre]:whitespace-pre",
        wordWrap && "[&>div>div>pre]:whitespace-pre-wrap",
        "min-w-full",
        className
      )}
    >
      {children}
    </div>
  </div>
));
Logs.displayName = "Logs";

export const LogEntry = React.forwardRef<HTMLDivElement, LogEntryProps>(
  ({ time, variant, children, className, ...props }, ref) => (
    <div
      ref={ref}
      {...props}
      className={twMergeClsx(
        "px-2 text-[13px] text-black dark:text-white",
        "flex min-w-full flex-row",
        variant === "error" &&
          "bg-danger-4 text-danger-10 dark:bg-danger-dark-4 dark:text-danger-dark-10",
        variant === "success" &&
          "bg-success-4 text-success-10 dark:bg-success-dark-4 dark:text-success-dark-10",
        variant === "warning" &&
          "bg-warning-4 text-warning-10 dark:bg-warning-dark-4 dark:text-warning-dark-10",
        variant === "info" &&
          "bg-info-4 text-info-10 dark:bg-info-dark-4 dark:text-info-dark-10",
        className
      )}
    >
      <div className="w-24 shrink-0 self-start pr-2 font-menlo">{time}</div>
      <div className="self-center font-menlo leading-5">
        <pre className="whitespace-pre-line">{children}</pre>
      </div>
    </div>
  )
);
LogEntry.displayName = "LogEntry";
