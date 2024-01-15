import { AlertTriangle, CheckCircle, Info, XCircle } from "lucide-react";
import React, { ElementRef, FC, HTMLAttributes } from "react";

import { twMergeClsx } from "~/util/helpers";

export type AlertProps = HTMLAttributes<HTMLDivElement> & {
  variant?: "info" | "success" | "warning" | "error";
  ref?: React.ForwardedRef<HTMLDivElement>;
  children?: React.ReactNode;
};

const Alert: FC<AlertProps> = ({ variant, className, children, ...props }) => (
  <div
    className={twMergeClsx(
      "rounded-md p-2 shadow-sm",
      variant === "info" &&
        "bg-info-4 text-info-11 dark:bg-info-dark-4 dark:text-info-dark-11",
      variant === "error" &&
        "bg-danger-4 text-danger-11 dark:bg-danger-dark-4 dark:text-danger-dark-11",
      variant === "success" &&
        "bg-success-4 text-success-11 dark:bg-success-dark-4 dark:text-success-dark-11",
      variant === "warning" &&
        "bg-warning-4 text-warning-11 dark:bg-warning-dark-4 dark:text-warning-dark-11",
      variant === undefined && "bg-gray-2 dark:bg-gray-dark-2",
      className
    )}
    {...props}
  >
    <div className="flex flex-col items-center sm:flex-row">
      {variant === "success" && <CheckCircle />}
      {variant === "warning" && <AlertTriangle />}
      {variant === "info" && <Info />}
      {variant === "error" && <XCircle />}
      {variant === undefined && <Info />}
      <span className="flex-1 px-2">{children}</span>
    </div>
  </div>
);

const AlertWithForwaredRef = React.forwardRef<
  ElementRef<typeof Alert>,
  AlertProps
>(({ ...props }, ref) => <Alert ref={ref} {...props} />);

AlertWithForwaredRef.displayName = "Alert";

export default AlertWithForwaredRef;
