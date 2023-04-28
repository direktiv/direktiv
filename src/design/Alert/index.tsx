import { AlertTriangle, CheckCircle, Info, XCircle } from "lucide-react";
import React, { FC, HTMLAttributes } from "react";

import clsx from "clsx";

export type AlertProps = HTMLAttributes<HTMLDivElement> & {
  variant?: "info" | "success" | "warning" | "error";
  forwaredRef?: React.ForwardedRef<HTMLDivElement>;
  children?: React.ReactNode;
};

const Alert: FC<AlertProps> = ({ variant, className, children }) => (
  <div
    className={clsx(
      className,
      "ring-md rounded-md p-2 shadow-sm",
      variant === "info" &&
        "bg-info-4 text-info-9 dark:bg-info-dark-4 dark:text-info-dark-9",
      variant === "error" &&
        "bg-danger-4 text-danger-9 dark:bg-danger-dark-4 dark:text-danger-dark-9",
      variant === "success" &&
        "bg-success-4 text-success-9 dark:bg-success-dark-4 dark:text-success-dark-9",
      variant === "warning" &&
        "bg-warning-4 text-warning-9 dark:bg-warning-dark-4 dark:text-warning-dark-9",
      variant === undefined &&
        "bg-gray-4 text-gray-9 dark:bg-gray-dark-4 dark:text-gray-dark-9"
    )}
  >
    <div className="flex items-center [&>svg]:inline">
      {variant === "success" && <CheckCircle />}
      {variant === "warning" && <AlertTriangle />}
      {variant === "info" && <Info />}
      {variant === "error" && <XCircle />}
      {variant === undefined && <Info />}
      <span className="px-2">{children}</span>
    </div>
  </div>
);

const AlertWithForwaredRef = React.forwardRef<HTMLDivElement, AlertProps>(
  ({ ...props }, ref) => <Alert forwaredRef={ref} {...props} />
);

AlertWithForwaredRef.displayName = "Alert";

export default AlertWithForwaredRef;
