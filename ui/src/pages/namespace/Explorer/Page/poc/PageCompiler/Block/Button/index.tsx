import { DefaultButton, DefaultButtonProps } from "./DefaultButton";
import { Ref, forwardRef } from "react";
import { TextButton, TextButtonProps } from "./TextButton";

type ButtonCompoundProps =
  | ({ as?: "button" } & DefaultButtonProps)
  | ({ as: "text" } & TextButtonProps);

export const Button = forwardRef<HTMLElement, ButtonCompoundProps>(
  ({ as, ...props }, ref) => {
    if (as === "text") {
      return (
        <TextButton
          {...(props as TextButtonProps)}
          ref={ref as Ref<HTMLSpanElement>}
        />
      );
    }
    return (
      <DefaultButton
        {...(props as DefaultButtonProps)}
        ref={ref as Ref<HTMLButtonElement>}
      />
    );
  }
);

Button.displayName = "Button";
