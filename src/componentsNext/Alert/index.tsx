import React, { FC, HTMLAttributes } from "react";

import clsx from "clsx";

export type AlertProps = HTMLAttributes<HTMLDivElement> & {
  variant?: "info" | "success" | "warning" | "error";
  text?: string;
  className?: string;
  forwaredRef?: React.ForwardedRef<HTMLDivElement>;
  children?: React.ReactNode;
};

const Alert: FC<AlertProps> = ({ variant, className, text }) => (
  <div
    className={clsx(
      className,
      "alert shadow-lg",
      variant === "info" && "alert-info",
      variant === "error" && "alert-error",
      variant === "success" && "alert-success",
      variant === "warning" && "alert-warning"
    )}
  >
    <div>
      <svg
        xmlns="http://www.w3.org/2000/svg"
        fill="none"
        viewBox="0 0 24 24"
        className="h-6 w-6 flex-shrink-0 stroke-current"
      >
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth="2"
          d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
        ></path>
      </svg>
      <span>{text}</span>
    </div>
  </div>
);

const AlertWithForwaredRef = React.forwardRef<HTMLDivElement, AlertProps>(
  ({ ...props }, ref) => <Alert forwaredRef={ref} {...props} />
);

AlertWithForwaredRef.displayName = "Alert";

export default AlertWithForwaredRef;
