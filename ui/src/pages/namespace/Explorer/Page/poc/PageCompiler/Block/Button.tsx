import ButtonDesignComponent from "~/design/Button";
import { ButtonType } from "../../schema/blocks/button";
import { forwardRef } from "react";

type ButtonProps = {
  blockProps: ButtonType;
};

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ blockProps, ...props }, ref) => {
    // TODO: implement the submit
    const { label, submit: _submit } = blockProps;
    return (
      <ButtonDesignComponent ref={ref} {...props}>
        {label}
      </ButtonDesignComponent>
    );
  }
);

Button.displayName = "Button";
