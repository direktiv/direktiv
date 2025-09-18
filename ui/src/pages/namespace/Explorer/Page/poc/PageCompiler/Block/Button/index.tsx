import ButtonDesignComponent, {
  ButtonProps as ButtonDesignComponentProps,
} from "~/design/Button";

import { ButtonType } from "../../../schema/blocks/button";
import { TemplateString } from "../../primitives/TemplateString";
import { forwardRef } from "react";

export type DefaultButtonProps = React.ButtonHTMLAttributes<HTMLButtonElement> &
  ButtonDesignComponentProps & {
    blockProps: ButtonType;
  };

export const Button = forwardRef<HTMLButtonElement, DefaultButtonProps>(
  ({ blockProps, ...props }, ref) => (
    <ButtonDesignComponent ref={ref} {...props}>
      <TemplateString value={blockProps.label} />
    </ButtonDesignComponent>
  )
);

Button.displayName = "Button";
