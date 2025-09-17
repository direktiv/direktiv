import { DefaultButton, DefaultButtonProps } from "./DefaultButton";
import { Ref, forwardRef } from "react";
import { TextButton, TextButtonProps } from "./TextButton";

type ButtonCompoundProps =
  | ({ as?: "button"; disabled?: boolean } & DefaultButtonProps)
  // TODO: is this still needed?
  | ({ as: "text" } & TextButtonProps);

export const Button = forwardRef<
  HTMLButtonElement | HTMLSpanElement,
  ButtonCompoundProps
>(({ as, ...props }, ref) => {
  if (as === "text") {
    return <TextButton {...props} ref={ref as Ref<HTMLSpanElement>} />;
  }
  return <DefaultButton {...props} ref={ref as Ref<HTMLButtonElement>} />;
});

Button.displayName = "Button";
