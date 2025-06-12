import ButtonDesignComponent, {
  ButtonProps as ButtonDesignComponentProps,
} from "~/design/Button";

import { ButtonType } from "../../../schema/blocks/button";
import { forwardRef } from "react";

export type DefaultButtonProps = ButtonDesignComponentProps & {
  blockProps: ButtonType;
};

export const DefaultButton = forwardRef<HTMLButtonElement, DefaultButtonProps>(
  ({ blockProps, ...props }, ref) => (
    <ButtonDesignComponent ref={ref} {...props}>
      {blockProps.label}
    </ButtonDesignComponent>
  )
);
DefaultButton.displayName = "DefaultButton";
