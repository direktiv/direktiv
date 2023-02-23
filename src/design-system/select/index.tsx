import React, { FC } from "react";

const Select: FC<{
  size?: "xs" | "sm" | "lg";
  border?: boolean;
  ghost?: boolean;
  className?: string;
  children?: React.ReactNode;
}> = ({ size, border, ghost, className, children }) => {
  let sizeClass;
  switch (size) {
    case "lg":
      sizeClass = "select-lg";
      break;
    case "sm":
      sizeClass = "select-sm";
      break;
    case "xs":
      sizeClass = "select-xs";
      break;
  }

  const borderClass = border ? "select-bordered" : "";
  const ghostClass = ghost ? "select-ghost" : "";

  return (
    <select
      className={`select ${sizeClass} ${borderClass} ${ghostClass} ${
        className ?? ""
      }}`}
    >
      {children}
    </select>
  );
};

export default Select;
