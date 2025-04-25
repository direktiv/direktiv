import ButtonDesignComponent from "~/design/Button";
import { ButtonType } from "../../schema/blocks/button";
import { forwardRef } from "react";

type ButtonProps = {
  blockProps: ButtonType;
};

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  (
    {
      // TODO: implement the submit
      blockProps: { label, submit: _submit },
      ...props
    },
    ref
  ) => (
    <ButtonDesignComponent ref={ref} {...props}>
      {label}
    </ButtonDesignComponent>
  )
);

Button.displayName = "Button";
