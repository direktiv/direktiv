import CopyButton from "~/design/CopyButton";
import { FC } from "react";
import Input from "~/design/Input";
import { InputWithButton } from "~/design/InputWithButton";

type PublicPathInputProps = {
  path: string;
};

const PublicPathInput: FC<PublicPathInputProps> = ({ path }) => (
  <InputWithButton
    onClick={(e) => {
      e.stopPropagation(); // prevent the onClick on the row from firing when clicking the workflow link
    }}
  >
    <Input value={path} readOnly />
    <CopyButton
      value={path}
      buttonProps={{
        variant: "ghost",
        icon: true,
      }}
    />
  </InputWithButton>
);

export default PublicPathInput;
