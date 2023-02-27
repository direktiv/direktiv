import React, { FC } from "react";

import clsx from "clsx";

export type ButtonProps = {
  size?: "xs" | "sm" | "lg";
  color?:
    | "primary"
    | "secondary"
    | "accent"
    | "ghost"
    | "link"
    | "info"
    | "success"
    | "warning"
    | "error";
  outline?: boolean;
  active?: boolean;
  loading?: boolean;
  circle?: boolean;
  block?: boolean;
  className?: string;
  forwaredRef?: React.ForwardedRef<HTMLButtonElement>;
  children?: React.ReactNode;
};

const Button: FC<ButtonProps> = ({
  size,
  color,
  outline,
  active,
  loading,
  circle,
  block,
  className,
  children,
  forwaredRef,
  ...props
}) => (
  <button
    className={clsx(
      className,
      "btn gap-2",
      size === "lg" && "btn-lg gap-3 [&>svg]:h-6",
      !size && "gap-2 [&>svg]:h-6",
      size === "sm" && "btn-sm gap-1 [&>svg]:h-5",
      size === "xs" && "btn-xs gap-0.5 [&>svg]:h-4",
      color === "primary" && "btn-primary",
      color === "secondary" && "btn-secondary",
      color === "accent" && "btn-accent",
      color === "ghost" && "btn-ghost",
      color === "link" && "btn-link",
      color === "info" && "btn-info",
      color === "success" && "btn-success",
      color === "warning" && "btn-warning",
      color === "error" && "btn-error",
      active && "btn-active",
      outline && "btn-outline",
      loading && "loading",
      circle && "btn-circle",
      block && "btn-block"
    )}
    ref={forwaredRef}
    {...props}
  >
    {children}
  </button>
);

const ButtonWithForwaredRef = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ ...props }, ref) => <Button forwaredRef={ref} {...props} />
);

ButtonWithForwaredRef.displayName = "Button";

export default ButtonWithForwaredRef;
