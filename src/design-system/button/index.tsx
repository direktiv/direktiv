import React, { FC } from "react";
import clsx from "clsx";

const Button: FC<{
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
  children?: React.ReactNode;
}> = ({
  size,
  color,
  outline,
  active,
  loading,
  circle,
  block,
  className,
  children,
  ...props
}) => (
  <button
    className={clsx(
      className,
      "btn",
      size === "lg" && "btn-lg",
      size === "sm" && "btn-sm",
      size === "xs" && "btn-xs",
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
    {...props}
  >
    {children}
  </button>
);

export default Button;
