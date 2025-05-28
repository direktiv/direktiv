import ButtonDesignComponent, {
  ButtonProps as ButtonDesignComponentProps,
} from "~/design/Button";
import { HTMLAttributes, forwardRef } from "react";

import { ButtonType } from "../../schema/blocks/button";

/**
 * BUTTON
 */
type ButtonAProps = ButtonDesignComponentProps & {
  blockProps: ButtonType;
};

const ButtonA = forwardRef<HTMLButtonElement, ButtonAProps>(
  ({ blockProps, ...props }, ref) => (
    <ButtonDesignComponent ref={ref} {...props}>
      {blockProps.label}
    </ButtonDesignComponent>
  )
);
ButtonA.displayName = "Button";

/**
 * SPAN
 */
type ButtonSpanProps = HTMLAttributes<HTMLSpanElement> & {
  blockProps: ButtonType;
};

const ButtonSpan = forwardRef<HTMLSpanElement, ButtonSpanProps>(
  ({ blockProps, ...props }, ref) => (
    <span ref={ref} {...props}>
      {blockProps.label}
    </span>
  )
);
ButtonSpan.displayName = "Span";

type ButtonCompoundProps =
  | ({ as?: "button" } & ButtonAProps)
  | ({ as: "span" } & ButtonSpanProps);

export const Button = forwardRef<HTMLElement, ButtonCompoundProps>(
  ({ as, ...props }, ref) => {
    if (as === "span") {
      return (
        <ButtonSpan
          {...(props as ButtonSpanProps)}
          ref={ref as React.Ref<HTMLSpanElement>}
        />
      );
    }
    return (
      <ButtonA
        {...(props as ButtonAProps)}
        ref={ref as React.Ref<HTMLButtonElement>}
      />
    );
  }
);

Button.displayName = "Button";
