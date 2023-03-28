import React, { FC, HTMLAttributes } from "react";

import clsx from "clsx";

export type TextInputProps = HTMLAttributes<HTMLDivElement> & {
  className?: string;
  variant?:
    | "primary"
    | "secondary"
    | "accent"
    | "ghost"
    | "success"
    | "warning"
    | "error"
    | "info";
  size?: "xs" | "sm" | "lg";
  block?: boolean;
  forwaredRef?: React.ForwardedRef<HTMLDivElement>;
  children?: React.ReactNode;
};

const TextInput: FC<TextInputProps> = ({
  className,
  variant,
  size,
  block = false,
  ...props
}) => (
  <input
    type="text"
    className={clsx(
      className,
      "input",
      size === "lg" && "input-lg",
      size === "sm" && "input-sm",
      size === "xs" && "input-xs",
      variant === "primary" && "input-primary",
      variant === "secondary" && "input-secondary",
      variant === "accent" && "input-accent",
      variant === "info" && "input-info",
      variant === "success" && "input-success",
      variant === "warning" && "input-warning",
      variant === "error" && "input-error",
      variant === "ghost" ? "input-ghost" : "input-bordered",
      block && "w-full"
    )}
    {...props}
  />
);

const TextInputWithForwaredRef = React.forwardRef<
  HTMLDivElement,
  TextInputProps
>(({ ...props }, ref) => <TextInput forwaredRef={ref} {...props} />);

TextInputWithForwaredRef.displayName = "TextInput";

export default TextInputWithForwaredRef;
