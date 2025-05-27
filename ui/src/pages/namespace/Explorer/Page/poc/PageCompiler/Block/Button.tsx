import ButtonDesignComponent, {
  ButtonProps as ButtonDesignComponentProps,
} from "~/design/Button";

import { ButtonType } from "../../schema/blocks/button";
import { forwardRef } from "react";

type SpanButtonProps = {
  as: "span";
} & React.HTMLAttributes<HTMLSpanElement>;

type DefaultButtonProps = {
  as?: "button";
} & React.HTMLAttributes<HTMLButtonElement> &
  ButtonDesignComponentProps;

type ButtonProps<T extends "span" | "button" = "button"> = {
  blockProps: ButtonType;
} & (T extends "span" ? SpanButtonProps : DefaultButtonProps);

export const Button = forwardRef<
  HTMLElement,
  ButtonProps<"span"> | ButtonProps<"button">
>(({ as, blockProps, ...props }, ref) => {
  const { label } = blockProps;

  if (as === "span") {
    return (
      <span ref={ref as React.Ref<HTMLSpanElement>} {...props}>
        {label}
      </span>
    );
  }

  return (
    <ButtonDesignComponent ref={ref as React.Ref<HTMLButtonElement>} {...props}>
      {label}
    </ButtonDesignComponent>
  );
});

Button.displayName = "Button";
