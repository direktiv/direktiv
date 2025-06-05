import { HTMLAttributes, forwardRef } from "react";

import { ButtonType } from "../../../schema/blocks/button";

export type TextButtonProps = HTMLAttributes<HTMLSpanElement> & {
  blockProps: ButtonType;
};

export const TextButton = forwardRef<HTMLSpanElement, TextButtonProps>(
  ({ blockProps, ...props }, ref) => (
    <span ref={ref} {...props}>
      {blockProps.label}
    </span>
  )
);
TextButton.displayName = "TextButton";
