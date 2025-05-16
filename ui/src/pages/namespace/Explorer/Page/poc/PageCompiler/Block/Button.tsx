import { ElementType, forwardRef } from "react";

import ButtonDesignComponent from "~/design/Button";
import { ButtonType } from "../../schema/blocks/button";

type AsComponent = ElementType | typeof ButtonDesignComponent;

interface ButtonProps {
  as?: AsComponent;
  blockProps: ButtonType;
}

export const Button = forwardRef<HTMLElement, ButtonProps>(
  ({ as, blockProps, ...props }, ref) => {
    const { label } = blockProps;
    const Component = as || ButtonDesignComponent;
    return (
      <Component ref={ref} {...props}>
        {label}
      </Component>
    );
  }
);

Button.displayName = "Button";
