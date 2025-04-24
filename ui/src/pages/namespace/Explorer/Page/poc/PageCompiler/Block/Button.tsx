import ButtonDesignComponent from "~/design/Button";
import { ButtonType } from "../../schema/blocks/button";

type ButtonProps = {
  blockProps: ButtonType;
};

export const Button = ({
  // TODO: implement the submit
  blockProps: { label, submit: _submit },
}: ButtonProps) => <ButtonDesignComponent>{label}</ButtonDesignComponent>;
