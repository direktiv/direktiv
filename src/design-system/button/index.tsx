import React, { FC } from "react";

const Button: FC<{
  size?: "xs" | "sm" | "lg";
  color?: "primary" | "secondary" | "accent" | "ghost" | "link";
  state?: "info" | "success" | "warning" | "error";
  active?: boolean;
  outline?: boolean;
  loading?: boolean;
  children?: React.ReactNode;
  className?: string;
}> = ({
  size,
  color,
  state,
  active,
  outline,
  loading,
  className,
  children,
}) => {
  let sizeClass;
  switch (size) {
    case "lg":
      sizeClass = "btn-lg";
      break;
    case "sm":
      sizeClass = "btn-sm";
      break;
    case "xs":
      sizeClass = "btn-xs";
      break;
  }

  let colorClass;
  switch (color) {
    case "primary":
      colorClass = "btn-primary";
      break;
    case "secondary":
      colorClass = "btn-secondary";
      break;
    case "accent":
      colorClass = "btn-accent";
      break;
    case "ghost":
      colorClass = "btn-ghost";
      break;
    case "link":
      colorClass = "btn-link";
      break;
  }

  let stateClass;
  switch (state) {
    case "info":
      stateClass = "btn-info";
      break;
    case "success":
      stateClass = "btn-success";
      break;
    case "warning":
      stateClass = "btn-warning";
      break;
    case "error":
      stateClass = "btn-error";
      break;
  }

  const activeClass = active ? "btn-active" : "";
  const outlineClass = outline ? "btn-outline" : "";
  const loadingClass = loading ? "loading" : "";

  return (
    <button
      className={`btn ${sizeClass} ${colorClass} ${stateClass} ${activeClass} ${outlineClass} ${loadingClass} ${
        className ?? ""
      }}`}
    >
      {children}
    </button>
  );
};

export default Button;
