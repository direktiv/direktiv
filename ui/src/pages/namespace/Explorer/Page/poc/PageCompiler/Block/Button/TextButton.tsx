import { HTMLAttributes, forwardRef } from "react";

import { ButtonType } from "../../../schema/blocks/button";
import { TemplateString } from "../../primitives/TemplateString";

export type TextButtonProps = HTMLAttributes<HTMLSpanElement> & {
  blockProps: ButtonType;
};

export const TextButton = forwardRef<HTMLSpanElement, TextButtonProps>(
  ({ blockProps, ...props }, ref) => (
    <span ref={ref} {...props}>
      <TemplateString value={blockProps.label} />
    </span>
  )
);
TextButton.displayName = "TextButton";
