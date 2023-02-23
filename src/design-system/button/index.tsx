import React, { FC } from "react";

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
  className?: string;
  children?: React.ReactNode;
}> = ({
  size,
  color,
  outline,
  active,
  loading,
  circle,
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
    case "info":
      colorClass = "btn-info";
      break;
    case "success":
      colorClass = "btn-success";
      break;
    case "warning":
      colorClass = "btn-warning";
      break;
    case "error":
      colorClass = "btn-error";
      break;
  }

  const activeClass = active ? "btn-active" : "";
  const outlineClass = outline ? "btn-outline" : "";
  const loadingClass = loading ? "loading" : "";
  const circleClass = circle ? "btn-circle" : "";

  return (
    <button
      className={`btn ${sizeClass} ${colorClass} ${activeClass} ${outlineClass} ${loadingClass} ${circleClass} ${
        className ?? ""
      }}`}
    >
      {children}
    </button>
  );
};

export default Button;
