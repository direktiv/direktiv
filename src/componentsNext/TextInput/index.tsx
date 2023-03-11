import React, { FC, HTMLAttributes } from "react";

import clsx from "clsx";

export type TextInputProps = HTMLAttributes<HTMLDivElement> & {
  className?: string;
  variant?:
    | "primary"
    | "secondary"
    | "accent"
    | "success"
    | "warning"
    | "error"
    | "info";
  size?: "xs" | "sm" | "md" | "lg";
  block?: boolean;
  ghost?: boolean;
  forwaredRef?: React.ForwardedRef<HTMLDivElement>;
  children?: React.ReactNode;
};

const TextInput: FC<TextInputProps> = ({
  className,
  variant,
  size = "lg",
  block = false,
  ghost = false,
  ...props
}) => (
  <input
    type="text"
    className={clsx(
      className,
      "input max-w-xs",
      ghost ? "input-ghost" : "input-bordered",
      size === "lg" && "input-lg",
      size === "md" && "input-md",
      size === "sm" && "input-sm",
      size === "xs" && "input-xs",
      variant === "primary" && "input-primary",
      variant === "secondary" && "input-secondary",
      variant === "accent" && "input-accent",
      variant === "info" && "input-info",
      variant === "success" && "input-success",
      variant === "warning" && "input-warning",
      variant === "error" && "input-error",
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
