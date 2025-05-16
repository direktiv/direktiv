import { ButtonHTMLAttributes, forwardRef } from "react";

import ButtonDesignComponent from "~/design/Button";
import { ButtonType } from "../../schema/blocks/button";

type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  blockProps: ButtonType;
};

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ blockProps, ...props }, ref) => {
    // TODO: implement the submit
    const { label } = blockProps;
    return (
      <ButtonDesignComponent ref={ref} {...props}>
        {label}
      </ButtonDesignComponent>
    );
  }
);

Button.displayName = "Button";
